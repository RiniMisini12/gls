package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rinimisini112/gls/operations"
	"github.com/rinimisini112/gls/structures"
	"github.com/rinimisini112/gls/tui"

	"github.com/olekukonko/tablewriter"
)

func printTable(files []structures.FileInfo, showHidden, preview bool) {
	table := tablewriter.NewWriter(os.Stdout)
	headers := []string{"Name", "User:Group", "Permissions", "Size", "Last Modified", "Type"}
	if preview {
		headers = append(headers, "Preview")
	}
	table.SetHeader(headers)

	colorMap := map[string]tablewriter.Colors{
		"#FF0000": {tablewriter.FgRedColor},
		"#FFA500": {tablewriter.FgYellowColor},
		"#00FF00": {tablewriter.FgGreenColor},
	}

	table.SetColumnColor(
		tablewriter.Colors{tablewriter.FgWhiteColor},
		tablewriter.Colors{tablewriter.FgCyanColor},
		tablewriter.Colors{tablewriter.FgMagentaColor},
		tablewriter.Colors{},
		tablewriter.Colors{tablewriter.FgBlueColor},
		tablewriter.Colors{tablewriter.FgHiWhiteColor},
	)

	var coloredRows [][]string

	for _, file := range files {
		if !showHidden && file.Hidden {
			continue
		}

		row := []string{
			file.Name,
			file.UserAndGroup,
			file.Permissions,
			file.Size,
			file.ModTime,
			func() string {
				if file.IsDir {
					return "📂 Dir"
				}
				return "📄 File"
			}(),
		}

		if preview && !file.IsDir {
			row = append(row, operations.PreviewFile(file.Path, 3))
		}

		coloredRows = append(coloredRows, row)
	}

	for i, row := range coloredRows {
		sizeColor := colorMap[files[i].Color][0]
		row[3] = fmt.Sprintf("\x1b[%dm%s\x1b[0m", sizeColor+30, row[3])

		table.Rich(row, []tablewriter.Colors{
			{},
			{},
			{},
			{colorMap[files[i].Color][0]},
			{},
			{},
		})
	}

	table.Render()
}

func showHelp() {
	fmt.Println("\n📂 Usage: gls [options] [directories]")
	fmt.Println("Options:")
	fmt.Println("  -s=name      Sort by name (default)")
	fmt.Println("  -s=size      Sort by size")
	fmt.Println("  -s=date      Sort by date")
	fmt.Println("  -a           Show hidden files")
	fmt.Println("  -p           Preview files")
	fmt.Println("  -t=dir       Show only directories")
	fmt.Println("  -t=file      Show only files")
	fmt.Println("  -t=hidden    Show only hidden files")
	fmt.Println("  -l [limit]   Limit the number of files displayed")
	fmt.Println("  -s [query]   Search for files containing 'query'")
	fmt.Println("  --help, -h   Show this help message")
}

func main() {
	var dirs []string
	sortBy := "name"
	showHidden := false
	searchQuery := ""
	filterType := ""
	preview := false
	limit := -1
	interactive := false

	args := os.Args[1:]
	validArgs := map[string]bool{
		"--help": true, "-h": true,
		"--rename": true,
		"-s=name":  true, "-s=size": true, "-s=date": true,
		"-a": true, "-p": true,
		"-t=dir": true, "-t=file": true, "-t=hidden": true,
		"-l": true, "-s": true,
		"-i":            true,
		"--interactive": true,
		"--version":     true,
		"-v":            true,
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
		case arg == "-p":
			preview = true
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
			var filteredFiles []structures.FileInfo
			for _, file := range files {
				if strings.Contains(strings.ToLower(file.Name), strings.ToLower(searchQuery)) {
					filteredFiles = append(filteredFiles, file)
				}
			}
			files = filteredFiles
		}

		files = operations.Paginate(files, limit)

		printTable(files, showHidden, preview)
	}
}
