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

	if ev.Key() == tcell.KeyCtrlD {
		e.mode = "command"
		e.commandBuffer = "delete "
		e.SetStatusMessage("Enter filename to delete: ")
		return
	}

	switch e.mode {
	case "normal":
		e.handleNormalMode(ev)
	case "insert":
		e.handleInsertMode(ev)
	case "command":
		switch ev.Key() {
		case tcell.KeyEnter:
			e.handleCommand()
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if len(e.commandBuffer) > 0 {
				e.commandBuffer = e.commandBuffer[:len(e.commandBuffer)-1]
				e.SetStatusMessage(fmt.Sprintf(":%s", e.commandBuffer))
			}
		case tcell.KeyRune:
			e.commandBuffer += string(ev.Rune())
			e.SetStatusMessage(fmt.Sprintf(":%s", e.commandBuffer))
		}
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
	button := ev.Buttons()

	// Check if the click is within the file tree
	if e.treeVisible && x <= e.treeWidth {
		e.handleTreeMouseEvent(x, y, button)
		return
	}

	// Adjust coordinates for line numbers and tree
	adjustedX := x
	if e.showLineNumbers {
		adjustedX -= 5
	}
	if e.treeVisible {
		adjustedX -= e.treeWidth + 1
	}

	// Adjust y coordinate for scrolling
	adjustedY := y + e.scrollY

	// Handle clicks within the text area
	if adjustedX >= 0 && adjustedY >= 0 && adjustedY < len(e.lines) {
		e.cursorX = adjustedX
		e.cursorY = adjustedY

		// Ensure cursor stays within line bounds
		if e.cursorX > len(e.lines[e.cursorY]) {
			e.cursorX = len(e.lines[e.cursorY])
		}

		// Handle different mouse buttons
		switch button {
		case tcell.Button1: // Left click
			// Do something on left click
			e.SetStatusMessage(fmt.Sprintf("Left click at x: %d, y: %d", adjustedX, adjustedY))
		case tcell.Button2: // Middle click
			// Do something on middle click
			e.SetStatusMessage("Middle click")
		case tcell.Button3: // Right click
			// Do something on right click
			e.SetStatusMessage("Right click")
		case tcell.WheelUp: // Mouse wheel up
			e.scrollUp()
		case tcell.WheelDown: // Mouse wheel down
			e.scrollDown()
		}
	}
}

func (e *Editor) scrollUp() {
	if e.scrollY > 0 {
		e.scrollY--
	}
}

func (e *Editor) scrollDown() {
	if e.scrollY < len(e.lines)-1 {
		e.scrollY++
	}
}

func (e *Editor) handleTreeMouseEvent(x, y int, button tcell.ButtonMask) {
	// Adjust y coordinate for scrolling
	adjustedY := y + e.scrollY

	// Check if the click is within the tree bounds
	if adjustedY >= 0 && adjustedY < len(e.fileTree.children) {
		e.treeSelectedLine = adjustedY
		e.SetStatusMessage(fmt.Sprintf("Selected line in tree: %d", adjustedY))

		// Handle different mouse buttons
		switch button {
		case tcell.Button1: // Left click
			// Open the selected file or directory
			selectedNode := e.fileTree.children[e.treeSelectedLine]
			if selectedNode.isDir {
				e.currentPath = selectedNode.name
				e.initFileTree()
			} else {
				e.LoadFile(selectedNode.name)
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
