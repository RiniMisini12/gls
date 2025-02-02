package tui

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mholt/archiver/v3"
	"github.com/rinimisini112/gls/operations"
	"github.com/rinimisini112/gls/structures"
	"github.com/rivo/tview"
)

type UIState struct {
	App        *tview.Application
	FileList   *tview.List
	Preview    *tview.TextView
	Pages      *tview.Pages
	CurrentDir string
	Files      []structures.FileInfo
	Selected   map[int]struct{}
}

func StartInteractiveMode(dir string) {
	app := tview.NewApplication()
	state := &UIState{
		App:        app,
		CurrentDir: dir,
		Selected:   make(map[int]struct{}),
	}

	flex := tview.NewFlex().
		AddItem(createFileList(state), 0, 1, true).
		AddItem(createPreviewPane(state), 40, 1, false)

	state.Pages = tview.NewPages().
		AddPage("main", flex, true, true)

	app.SetRoot(state.Pages, true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'j', 'J':
			state.FileList.SetCurrentItem((state.FileList.GetCurrentItem() + 1) % len(state.Files))
		case 'k', 'K':
			idx := state.FileList.GetCurrentItem() - 1
			if idx < 0 {
				idx = len(state.Files) - 1
			}
			state.FileList.SetCurrentItem(idx)
		case 'h', 'H':
			navigateUp(state)
		case 'l', 'L':
			enterDirectory(state)
		case 'e':
			openEditor(state)
		case 'd':
			deleteFile(state)
		case 's':
			showStats(state)
		case ' ':
			toggleSelection(state)
		case 'a':
			toggleAll(state)
		case 'A':
			createArchive(state)
		case 'q':
			state.App.Stop()
		}
		return event
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func createFileList(state *UIState) *tview.List {
	list := tview.NewList().ShowSecondaryText(false)
	state.FileList = list

	files, _ := operations.ListFiles(state.CurrentDir, "name")
	state.Files = files

	for _, file := range files {
		icon := "📄"
		if file.IsDir {
			icon = "📂"
		}
		list.AddItem(fmt.Sprintf("%s %s", icon, file.Name), "", 0, nil)
	}

	return list
}

func createPreviewPane(state *UIState) *tview.TextView {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(true)
	state.Preview = textView

	return textView
}

func navigateUp(state *UIState) {
	if state.CurrentDir == "/" {
		return
	}

	parentDir := filepath.Dir(state.CurrentDir)
	files, _ := operations.ListFiles(parentDir, "name")
	state.CurrentDir = parentDir
	state.Files = files
	state.FileList.Clear()

	for _, file := range files {
		icon := "📄"
		if file.IsDir {
			icon = "📂"
		}
		state.FileList.AddItem(fmt.Sprintf("%s %s", icon, file.Name), "", 0, nil)
	}
}

func enterDirectory(state *UIState) {
	currentSelection := state.FileList.GetCurrentItem()
	if currentSelection >= len(state.Files) {
		return
	}

	file := state.Files[currentSelection]
	if file.IsDir {
		newDir := file.Path
		files, _ := operations.ListFiles(newDir, "name")
		state.CurrentDir = newDir
		state.Files = files
		state.FileList.Clear()
		for _, file := range files {
			icon := "📄"
			if file.IsDir {
				icon = "📂"
			}
			state.FileList.AddItem(fmt.Sprintf("%s %s", icon, file.Name), "", 0, nil)
		}
	}
}

func openEditor(state *UIState) {
	currentSelection := state.FileList.GetCurrentItem()
	if currentSelection >= len(state.Files) {
		return
	}

	file := state.Files[currentSelection]
	if !file.IsDir {
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "nano"
		}
		cmd := exec.Command(editor, file.Path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
}

func deleteFile(state *UIState) {
	currentSelection := state.FileList.GetCurrentItem()
	if currentSelection >= len(state.Files) {
		return
	}

	file := state.Files[currentSelection]
	if !file.IsDir {
		confirm := confirmAction(fmt.Sprintf("Delete %s? (y/n): ", file.Name))
		if confirm {
			if err := os.Remove(file.Path); err != nil {
				log.Fatal(err)
			}
			state.FileList.RemoveItem(currentSelection)
		}
	}
}

func showStats(state *UIState) {
	currentSelection := state.FileList.GetCurrentItem()
	if currentSelection >= len(state.Files) {
		return
	}

	file := state.Files[currentSelection]
	stats, err := os.Stat(file.Path)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("File: %s\n", file.Name)
	fmt.Printf("Size: %d bytes\n", stats.Size())
	fmt.Printf("Permissions: %s\n", stats.Mode())
	fmt.Printf("Last modified: %v\n", stats.ModTime())
}

func toggleSelection(state *UIState) {
	currentSelection := state.FileList.GetCurrentItem()
	if currentSelection >= len(state.Files) {
		return
	}

	if _, ok := state.Selected[currentSelection]; ok {
		delete(state.Selected, currentSelection)
	} else {
		state.Selected[currentSelection] = struct{}{}
	}
}

func toggleAll(state *UIState) {
	if len(state.Selected) == len(state.Files) {
		state.Selected = make(map[int]struct{})
	} else {
		for i := range state.Files {
			state.Selected[i] = struct{}{}
		}
	}
}

func createArchive(state *UIState) {
	if len(state.Selected) == 0 {
		return
	}

	archiveName := fmt.Sprintf("%s.zip", filepath.Base(state.CurrentDir))
	var filesToArchive []string

	for idx := range state.Selected {
		filesToArchive = append(filesToArchive, state.Files[idx].Path)
	}

	err := archiver.Archive(filesToArchive, archiveName)
	if err != nil {
		fmt.Println("❌ Archive Error:", err)
	} else {
		fmt.Printf("✅ Archive Created: %s\n", archiveName)
	}
}

func confirmAction(msg string) bool {
	fmt.Print(msg)
	var input string
	fmt.Scanln(&input)
	return strings.ToLower(input) == "y"
}
