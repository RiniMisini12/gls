package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
	"github.com/rinimisini112/gls/finder"
	"github.com/rinimisini112/gls/operations"
	"github.com/rinimisini112/gls/structures"
	"github.com/rinimisini112/gls/tui"
)

func printTable(files []structures.FileInfo, showHidden bool, withGroupAndUser bool, fullDirSize bool) {
	if len(files) == 0 {
		fmt.Println("")
		fmt.Println("All who wander are not lost, But what you are looking for is nowhere to be found")
		return
	}

	table := tablewriter.NewWriter(os.Stdout)
	headers := []string{"Type", "Name"}
	if withGroupAndUser {
		headers = append(headers, "User:Group")
	}
	headers = append(headers, "Permissions", "Size", "Last Modified")

	table.SetHeader(headers)
	table.SetRowLine(true)
	table.SetAutoWrapText(false)

	for _, file := range files {
		if !showHidden && file.Hidden {
			continue
		}

		row := []string{
			func() string {
				if file.IsDir {
					return "📂 Dir"
				}
				return "📄 File"
			}(),
			file.Name,
		}

		if withGroupAndUser {
			row = append(row, file.UserAndGroup)
		}

		row = append(row, file.Permissions)

		size := file.RawSize
		if fullDirSize && file.IsDir {
			size = finder.CalculateDirSize(file.Path)
		}

		colorSize := func(size int64) string {
			sizeStr := humanize.Bytes(uint64(size))
			sizeStr = strings.ReplaceAll(sizeStr, "\n", "")
			sizeStr = strings.ReplaceAll(sizeStr, "\u00A0", " ")
			sizeStr = strings.TrimSpace(sizeStr)

			switch {
			case size < 1<<20:
				return "\033[32m" + sizeStr + "\033[0m"
			case size < 50<<20:
				return "\033[33m" + sizeStr + "\033[0m"
			case size < 500<<20:
				return "\033[38;5;208m" + sizeStr + "\033[0m"
			default:
				return "\033[31m" + sizeStr + "\033[0m"
			}
		}

		row = append(row, colorSize(size))
		row = append(row, file.ModTime)

		table.Append(row)
	}

	table.Render()
}

func showHelp() {
	fmt.Println("\n📂 Usage: gls [options] [directories]")
	fmt.Println("Options:")
	fmt.Println("  -s=name 	    Sort by name (default)")
	fmt.Println("  -s=size    	Sort by size")
	fmt.Println("  -s=date    	Sort by date")
	fmt.Println("  -a         	Show hidden files")
	fmt.Println("  -p         	Preview files")
	fmt.Println("  -t=dir     	Show only directories")
	fmt.Println("  -t=file    	Show only files")
	fmt.Println("  -t=hidden  	Show only hidden files")
	fmt.Println("  -l [limit] 	Limit the number of files displayed")
	fmt.Println("  -s [query] 	Search for files containing 'query'")
	fmt.Println("  --help, -h   Show this help message")
	fmt.Println("  --rename <old> <new>     Rename a file")
	fmt.Println("  -i           Interactive mode")
	fmt.Println("  -sa          Search for files with user and group")
	fmt.Println("  --version, -v  Show version")
	fmt.Println("  -fullDirSize  Show full directory size")
}

func main() {
	var dirs []string
	sortBy := "name"
	showHidden := false
	searchQuery := ""
	filterType := ""
	limit := -1
	interactive := false
	withGroupAndUser := false
	fullDirSize := false

	args := os.Args[1:]
	validArgs := map[string]bool{
		"--help": true, "-h": true,
		"--rename": true,
		"-s=name":  true, "-s=size": true, "-s=date": true,
		"-a": true, "-p": true,
		"-t=dir": true, "-t=file": true, "-t=hidden": true,
		"-l": true, "-s": true,
		"-i":            true,
		"-sa":           true,
		"--interactive": true,
		"--version":     true,
		"-v":            true,
		"-fullDirSize":  true,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if _, exists := validArgs[arg]; !exists && !strings.HasPrefix(arg, "-t=") {
			fmt.Println("❌ Invalid option:", arg)
			showHelp()
			return
		}

		switch {
		case arg == "--help" || arg == "-h":
			showHelp()
			return

		case arg == "--interactive" || arg == "-i":
			interactive = true

		case arg == "--rename":
			if i+2 >= len(args) {
				fmt.Println("❌ Error: --rename requires two arguments")
				fmt.Println("Usage: gsl --rename \"<old name>\" \"<new name>\"")
				return
			}

			oldName, newName, consumed := operations.ParseQuotedFilenames(args[i+1:])
			if consumed < 2 {
				fmt.Println("❌ Error: invalid arguments for --rename")
				fmt.Println("Usage: gsl --rename \"<old name>\" \"<new name>\"")
				return
			}

			if err := operations.Rename(oldName, newName); err != nil {
				fmt.Printf("❌ Rename error: %v\n", err)
				return
			}

			fmt.Printf("✅ Successfully renamed %q to %q\n", oldName, newName)
			i += consumed
			return
		case arg == "-s=size":
			sortBy = "size"
		case arg == "-s=date":
			sortBy = "date"
		case arg == "-a":
			showHidden = true
		case arg == "-fullDirSize":
			fullDirSize = true
		case strings.HasPrefix(arg, "-t="):
			filterType = strings.Split(arg, "=")[1]
		case arg == "-l":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &limit)
				i++
			}
		case arg == "-s":
			if i+1 < len(args) {
				searchQuery = args[i+1]
				i++
			}
		case arg == "-sa":
			if i+1 < len(args) {
				searchQuery = args[i+1]
				withGroupAndUser = true
				i++
			}
		case arg == "-v" || arg == "--version":
			fmt.Println("gls v1.0")
			return
		default:
			dirs = append(dirs, arg)
		}
	}

	if interactive {
		if len(dirs) == 0 {
			dirs = append(dirs, ".")
		}
		tui.StartInteractiveMode(dirs[0])
		return
	}

	if len(dirs) == 0 {
		dirs = append(dirs, ".")
	}

	for _, dir := range dirs {
		fmt.Printf("\n📂 Listing: %s\n", dir)

		files, err := operations.ListFiles(dir, sortBy)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		files = operations.FilterFiles(files, filterType)

		if searchQuery != "" {
			fmt.Println("🔍 Searching for:", searchQuery)
			files = finder.Search(searchQuery, dir, filterType, withGroupAndUser, fullDirSize)
		}

		files = operations.Paginate(files, limit)

		printTable(files, showHidden, withGroupAndUser, fullDirSize)
	}
}
