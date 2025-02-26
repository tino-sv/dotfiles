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
		e.handleCommand("command")
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

func (e *Editor) handleMouse(ev *tcell.EventMouse) {
	x, y := ev.Position()
	button := ev.Buttons()

	// Handle mouse wheel scrolling
	if button&tcell.WheelUp != 0 {
		e.scroll(-3)
		return
	}
	if button&tcell.WheelDown != 0 {
		e.scroll(3)
		return
	}

	// Handle left click
	if button&tcell.Button1 != 0 {
		// Check if click is in file tree area
		if e.treeVisible && x < e.treeWidth {
			e.handleTreeClick(x, y)
			return
		}

		// Adjust for line numbers and tree
		contentX := x
		if e.showLineNumbers {
			contentX -= 5
		}
		if e.treeVisible {
			contentX -= e.treeWidth + 1
		}

		// Set cursor position if within text area
		if contentX >= 0 && y >= 0 && y < len(e.lines) {
			line := e.lines[y]
			if contentX <= len(line) {
				e.cursorX = contentX
				e.cursorY = y
			} else if len(line) > 0 {
				e.cursorX = len(line)
				e.cursorY = y
			}
		}
	}
}

// Add scrolling support
func (e *Editor) scroll(amount int) {
	// Implement scrolling logic
	// For now, we'll just move the cursor
	newY := e.cursorY + amount
	if newY >= 0 && newY < len(e.lines) {
		e.cursorY = newY
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
