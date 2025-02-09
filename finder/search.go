package finder

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/dustin/go-humanize"
	lru "github.com/hashicorp/golang-lru"
	"github.com/rinimisini112/gls/structures"
)

var userCache *lru.Cache
var groupCache *lru.Cache
var fileCache *lru.Cache
var once sync.Once

func initCaches() {
	once.Do(func() {
		userCache, _ = lru.New(128)
		groupCache, _ = lru.New(128)
		fileCache, _ = lru.New(128)
	})
}

func getUserName(uid uint32) string {
	initCaches()
	uidStr := fmt.Sprintf("%d", uid)

	if name, found := userCache.Get(uidStr); found {
		return name.(string)
	}

	usr, err := user.LookupId(uidStr)
	if err != nil {
		userCache.Add(uidStr, "unknown")
		return "unknown"
	}

	userCache.Add(uidStr, usr.Username)
	return usr.Username
}

func getGroupName(gid uint32) string {
	initCaches()
	gidStr := fmt.Sprintf("%d", gid)

	if name, found := groupCache.Get(gidStr); found {
		return name.(string)
	}

	grp, err := user.LookupGroupId(gidStr)
	if err != nil {
		groupCache.Add(gidStr, "unknown")
		return "unknown"
	}

	groupCache.Add(gidStr, grp.Name)
	return grp.Name
}

func CalculateDirSize(dirPath string) int64 {
	var totalSize int64 = 0

	info, err := os.Stat(dirPath)
	if err != nil {
		fmt.Printf("❌ Error accessing %s: %v\n", dirPath, err)
		return 0
	}
	if !info.IsDir() {
		fmt.Printf("⚠️ Warning: %s is not a directory!\n", dirPath)
		return 0
	}

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("⚠️ Skipping file %s due to error: %v\n", path, err)
			return nil
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		fmt.Printf("❌ Error calculating size for %s: %v\n", dirPath, err)
		return 0
	}

	return totalSize
}

func searchFiles(
	rootDir,
	query string,
	results chan<- structures.FileInfo,
	wg *sync.WaitGroup,
	withGroupAndUser bool,
	fullDirSize bool,
) {
	defer wg.Done()

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		fullPath := filepath.Join(rootDir, entry.Name())

		info, err := entry.Info()
		if err != nil {
			continue
		}

		var userAndGroup string
		if withGroupAndUser {
			userAndGroup = fmt.Sprintf("%s:%s", getUserName(info.Sys().(*syscall.Stat_t).Uid), getGroupName(info.Sys().(*syscall.Stat_t).Gid))
		}

		if strings.Contains(strings.ToLower(entry.Name()), strings.ToLower(query)) {
			results <- structures.FileInfo{
				Name:  entry.Name(),
				Path:  fullPath,
				IsDir: entry.IsDir(),
				Size: func() string {
					if entry.IsDir() && fullDirSize {
						return humanize.Bytes(uint64(CalculateDirSize(fullPath)))
					}
					return humanize.Bytes(uint64(info.Size()))
				}(),
				ModTime: info.ModTime().Format(time.RFC822),
				UserAndGroup: func() string {
					if withGroupAndUser && userAndGroup != "" {
						return userAndGroup
					}
					return ""
				}(),
				Permissions: info.Mode().String(),
			}
		}

		if entry.IsDir() {
			wg.Add(1)
			go searchFiles(fullPath, query, results, wg, withGroupAndUser, fullDirSize)
		}
	}
}

func Search(query, startDir, filterType string, withUserAndGroup bool, fullDirSize bool) []structures.FileInfo {
	startTime := time.Now()
	initCaches()

	cacheKey := fmt.Sprintf("%s|%s|%s", query, startDir, filterType)

	if cachedResult, found := fileCache.Get(cacheKey); found {
		fmt.Println("✅ Returning cached search results")
		return cachedResult.([]structures.FileInfo)
	}

	results := make(chan structures.FileInfo, 100)
	var wg sync.WaitGroup

	wg.Add(1)
	go searchFiles(startDir, query, results, &wg, withUserAndGroup, fullDirSize)

	go func() {
		wg.Wait()
		close(results)
	}()

	var matches []structures.FileInfo
	for result := range results {
		if filterType == "dir" && !result.IsDir {
			continue
		}
		if filterType == "file" && result.IsDir {
			continue
		}
		matches = append(matches, result)
	}

	fileCache.Add(cacheKey, matches)
	fmt.Printf("🔍 Search took %s\n", time.Since(startTime))
	return matches
}
