package operations

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/rinimisini112/gls/structures"
)

func Rename(oldName, newName string) error {
	if _, err := os.Stat(oldName); os.IsNotExist(err) {
		return fmt.Errorf("source file %q not found", oldName)
	}

	if _, err := os.Stat(newName); err == nil {
		return fmt.Errorf("destination file %q already exists", newName)
	}

	if err := os.Rename(oldName, newName); err != nil {
		return fmt.Errorf("rename failed: %v", err)
	}

	return nil
}

func CreateArchive(files []structures.FileInfo, format string) error {
	archiveName := fmt.Sprintf("archive_%d.%s", time.Now().Unix(), format)
	file, err := os.Create(archiveName)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	for _, f := range files {
		if !f.Selected {
			continue
		}

		src, err := os.Open(f.Path)
		if err != nil {
			return err
		}
		defer src.Close()

		stat, err := src.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(stat)
		if err != nil {
			return err
		}

		header.Name = f.Name
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if _, err := io.Copy(writer, src); err != nil {
			return err
		}
	}
	return nil
}

func PreviewFile(path string, lines int) string {
	file, err := os.Open(path)
	if err != nil {
		return "(cannot preview)"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	content := ""
	count := 0
	for scanner.Scan() {
		content += scanner.Text() + " "
		count++
		if count >= lines {
			break
		}
	}
	return content
}

func FilterFiles(files []structures.FileInfo, filterType string) []structures.FileInfo {
	if filterType == "" {
		return files
	}

	var filtered []structures.FileInfo
	for _, file := range files {
		switch filterType {
		case "dir":
			if file.IsDir {
				filtered = append(filtered, file)
			}
		case "file":
			if !file.IsDir {
				filtered = append(filtered, file)
			}
		case "hidden":
			if file.Hidden {
				filtered = append(filtered, file)
			}
		}
	}
	return filtered
}

func Paginate(files []structures.FileInfo, limit int) []structures.FileInfo {
	if limit > 0 && len(files) > limit {
		return files[:limit]
	}
	return files
}

func ListFiles(dir string, sortBy string) ([]structures.FileInfo, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	type result struct {
		file structures.FileInfo
		err  error
	}

	ch := make(chan result, len(files))

	for _, entry := range files {
		go func(f os.DirEntry) {
			info, err := f.Info()
			if err != nil {
				ch <- result{err: err}
				return
			}

			stat, ok := info.Sys().(*syscall.Stat_t)
			if !ok {
				ch <- result{err: fmt.Errorf("failed to get file stats")}
				return
			}

			usr, err := user.LookupId(fmt.Sprintf("%d", stat.Uid))
			username := "unknown"
			if err == nil {
				username = usr.Username
			}

			grp, err := user.LookupGroupId(fmt.Sprintf("%d", stat.Gid))
			groupname := "unknown"
			if err == nil {
				groupname = grp.Name
			}

			file := structures.FileInfo{
				Name:         f.Name(),
				UserAndGroup: fmt.Sprintf("%s:%s", username, groupname),
				Permissions:  info.Mode().String(),
				Size:         strings.ReplaceAll(humanize.Bytes(uint64(info.Size())), " ", ""),
				RawSize:      info.Size(),
				ModTime:      info.ModTime().Format(time.RFC822),
				IsDir:        f.IsDir(),
				Hidden:       f.Name()[0] == '.',
				Path:         filepath.Join(dir, f.Name()),
			}

			ch <- result{file: file}
		}(entry)
	}

	var fileList []structures.FileInfo
	for range files {
		res := <-ch
		if res.err == nil {
			fileList = append(fileList, res.file)
		}
	}

	for i := range fileList {
		switch {
		case fileList[i].RawSize > 1<<30:
			fileList[i].Color = "#FF0000"
		case fileList[i].RawSize > 1<<20:
			fileList[i].Color = "#FFA500"
		default:
			fileList[i].Color = "#00FF00"
		}
	}

	switch sortBy {
	case "size":
		sort.Slice(fileList, func(i, j int) bool {
			return fileList[i].RawSize > fileList[j].RawSize
		})
	case "name":
		sort.Slice(fileList, func(i, j int) bool {
			return fileList[i].Name < fileList[j].Name
		})
	case "date":
		sort.Slice(fileList, func(i, j int) bool {
			t1, _ := time.Parse(time.RFC822, fileList[i].ModTime)
			t2, _ := time.Parse(time.RFC822, fileList[j].ModTime)
			return t1.After(t2)
		})
	}

	return fileList, nil
}

func ParseQuotedFilenames(args []string) (string, string, int) {
	var oldName, newName strings.Builder
	consumed := 0
	inQuote := false
	current := &oldName

	for i, arg := range args {
		consumed++

		if strings.HasPrefix(arg, "\"") {
			inQuote = true
			arg = arg[1:]
		}

		if strings.HasSuffix(arg, "\"") {
			inQuote = false
			arg = arg[:len(arg)-1]
		}

		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(arg)

		if !inQuote {
			if current == &oldName {
				current = &newName
				inQuote = false
				continue
			}
			return oldName.String(), newName.String(), consumed
		}

		if i == len(args)-1 {
			return oldName.String(), newName.String(), consumed
		}
	}

	return oldName.String(), newName.String(), consumed
}
