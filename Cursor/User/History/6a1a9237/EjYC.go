package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// Handle all input-related functions
func (e *Editor) handleInput(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyEscape {
		if e.mode == "command" || e.mode == "search" || e.mode == "filename" || e.mode == "rename" || e.mode == "confirm" {
			e.mode = "normal"
			e.commandBuffer = ""
			e.searchTerm = ""
			e.newFileDir = ""
			e.confirmAction = nil
			e.SetStatusMessage("NORMAL")
		} else if e.mode == "insert" {
			e.mode = "normal"
			e.SetStatusMessage("NORMAL")
		}
		return
	}

	switch e.mode {
	case "normal":
		e.handleNormalMode(ev)
	case "insert":
		e.handleInsertMode(ev)
	case "command":
		e.handleCommandMode(ev)
	case "search":
		e.handleSearchMode(ev)
	case "filename":
		e.handleFilenameMode(ev)
	case "rename":
		e.handleRenameMode(ev)
	case "confirm":
		e.handleConfirmMode(ev)
	}
}

func (e *Editor) handleNormalMode(ev *tcell.EventKey) {
	if e.treeVisible {
		switch ev.Key() {
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'j', 'k', 'h', 'l', 'n', 'd', 'r':
				e.handleTreeNavigation(ev)
				return
			case 't': // Toggle file tree
				e.treeVisible = !e.treeVisible
				return
			}
		case tcell.KeyEnter:
			e.handleTreeNavigation(ev)
			return
		}
	}

	// Regular editor bindings when tree is not visible
	switch ev.Key() {
	case tcell.KeyRune:
		switch ev.Rune() {
		case 'i':
			e.mode = "insert"
			e.SetStatusMessage("-- INSERT MODE -- (Tab for completions, Esc to exit)")
		case ':':
			e.mode = "command"
			e.commandBuffer = ""
			e.SetStatusMessage("Enter command (:w = save, :q = quit, :wq = save and quit)")
		case '/':
			e.mode = "search"
			e.searchTerm = ""
			e.SetStatusMessage("Enter search term (Esc to cancel)")
		case 'h':
			e.moveCursor(-1, 0)
		case 'l':
			e.moveCursor(1, 0)
		case 'j':
			e.moveCursor(0, 1)
		case 'k':
			e.moveCursor(0, -1)
		case 'u':
			e.undo()
		case 'r':
			e.redo()
		case 'n':
			e.nextMatch()
		case 'N':
			e.previousMatch()
		case 't': // Toggle file tree
			e.treeVisible = !e.treeVisible
			if e.treeVisible {
				e.SetStatusMessage("File tree: 'n' new file, 'D' delete, 'r' rename, Enter to open")
			}
		case '?':
			e.showHelp()
			e.SetStatusMessage("Press any key to exit help")
		}
	}
}

func (e *Editor) handleInsertMode(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEscape:
		e.mode = "normal"
		e.SetStatusMessage("-- NORMAL MODE --")
	case tcell.KeyTab:
		completions := e.getCompletions()
		if len(completions) > 0 {
			// Insert the first completion
			line := e.lines[e.cursorY]
			start := e.cursorX
			for start > 0 && isIdentChar(rune(line[start-1])) {
				start--
			}
			prefix := line[start:e.cursorX]
			completion := completions[0].Text

			// Only insert the part of completion that's not already typed
			if len(prefix) > 0 && strings.HasPrefix(completion, prefix) {
				completion = completion[len(prefix):]
			}

			// Insert the completion
			e.lines[e.cursorY] = line[:e.cursorX] + completion + line[e.cursorX:]
			e.cursorX += len(completion)
			e.isDirty = true

			e.SetStatusMessage(fmt.Sprintf("Completed: %s", completions[0].Text))
			e.showCompletions()
		}
		return
	case tcell.KeyEnter:
		e.insertNewLine()
		e.SetStatusMessage("-- INSERT MODE --")
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if e.cursorX > 0 {
			e.deleteChar()
		} else if e.cursorY > 0 {
			e.joinLines()
		}
	case tcell.KeyRune:
		e.insertRune(ev.Rune())
	}
}

func (e *Editor) handleFilenameMode(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEnter:
		if e.commandBuffer != "" {
			newPath := filepath.Join(e.newFileDir, e.commandBuffer)
			f, err := os.Create(newPath)
			if err != nil {
				e.SetStatusMessage(fmt.Sprintf("Error creating file: %v", err))
			} else {
				f.Close()
				e.SetFilename(newPath)
				e.lines = []string{""}
				e.cursorX = 0
				e.cursorY = 0
				e.isDirty = false
				e.mode = "normal"
				e.treeVisible = false
				e.refreshFileTree()
				e.SetStatusMessage(fmt.Sprintf("Created new file: %s", newPath))
			}
		}
		e.mode = "normal"
		e.commandBuffer = ""
		e.newFileDir = ""
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(e.commandBuffer) > 0 {
			e.commandBuffer = e.commandBuffer[:len(e.commandBuffer)-1]
			e.SetStatusMessage(fmt.Sprintf("New file name: %s", e.commandBuffer))
		}
	case tcell.KeyRune:
		e.commandBuffer += string(ev.Rune())
		e.SetStatusMessage(fmt.Sprintf("New file name: %s", e.commandBuffer))
	}
}

func (e *Editor) handleRenameMode(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEnter:
		if e.commandBuffer != "" {
			oldPath := e.newFileDir // Original filename
			newPath := filepath.Join(filepath.Dir(oldPath), e.commandBuffer)
			err := os.Rename(oldPath, newPath)
			if err != nil {
				e.SetStatusMessage(fmt.Sprintf("Error renaming file: %v", err))
			} else {
				e.SetStatusMessage(fmt.Sprintf("Renamed to %s", newPath))
				e.refreshFileTree()
			}
		}
		e.mode = "normal"
		e.commandBuffer = ""
		e.newFileDir = ""
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(e.commandBuffer) > 0 {
			e.commandBuffer = e.commandBuffer[:len(e.commandBuffer)-1]
			e.SetStatusMessage(fmt.Sprintf("New name: %s", e.commandBuffer))
		}
	case tcell.KeyRune:
		e.commandBuffer += string(ev.Rune())
		e.SetStatusMessage(fmt.Sprintf("New name: %s", e.commandBuffer))
	}
}

func (e *Editor) handleConfirmMode(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyRune {
		switch ev.Rune() {
		case 'y', 'Y':
			if e.confirmAction != nil {
				e.confirmAction()
			}
			e.mode = "normal"
		case 'n', 'N':
			e.SetStatusMessage("Operation cancelled")
			e.mode = "normal"
		}
	}
}

func (e *Editor) handleMouseEvent(ev *tcell.EventMouse) {
	x, y := ev.Position()

	switch ev.Buttons() {
	case tcell.ButtonPrimary:
		if e.treeVisible && x < e.treeWidth {
			// Handle click in file tree
			e.handleTreeClick(x, y)
		} else {
			// Handle click in content area
			e.handleContentClick(x, y)
		}
	}
}

// Handle clicks in the file tree
func (e *Editor) handleTreeClick(x, y int) {
	// Find which node was clicked
	var clickedNode *FileNode

	// Count visible nodes to find which one was clicked
	e.treeSelectedLine = y
	clickedNode = e.getSelectedNode()

	if clickedNode != nil {
		if clickedNode.isDir {
			clickedNode.expanded = !clickedNode.expanded
			if clickedNode.expanded && len(clickedNode.children) == 0 {
				e.loadDirectory(clickedNode)
			}
		} else {
			e.SetFilename(clickedNode.name)
			if err := e.LoadFile(clickedNode.name); err == nil {
				// Keep tree visible for now
			}
		}
	}
}

// Add this new function to handle content area clicks
func (e *Editor) handleContentClick(x, y int) {
	// Adjust y position based on scroll
	actualY := y + e.scrollY

	// Ensure click is within valid range
	if actualY >= 0 && actualY < len(e.lines) {
		// Calculate x position accounting for line numbers and tree if visible
		actualX := x
		if e.showLineNumbers {
			actualX -= 5 // Adjust for line number width
		}
		if e.treeVisible {
			actualX -= e.treeWidth + 1 // Adjust for tree width and separator
		}

		// Ensure x position is valid
		if actualX < 0 {
			actualX = 0
		}
		lineLength := len(e.lines[actualY])
		if actualX > lineLength {
			actualX = lineLength
		}

		// Update cursor position
		e.cursorY = actualY
		e.cursorX = actualX

		// Update scroll position if needed
		if e.cursorY < e.scrollY {
			e.scrollY = e.cursorY
		} else if e.cursorY >= e.scrollY+e.screenHeight-2 {
			e.scrollY = e.cursorY - (e.screenHeight - 3)
		}
	}
}
