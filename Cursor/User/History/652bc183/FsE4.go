package editor

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// Command mode functionality
func (e *Editor) handleCommandMode(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEnter:
		e.handleCommand()
		e.mode = "normal"
		e.commandBuffer = ""
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(e.commandBuffer) > 0 {
			e.commandBuffer = e.commandBuffer[:len(e.commandBuffer)-1]
			e.setStatusMessage(":" + e.commandBuffer)
		}
	case tcell.KeyRune:
		e.commandBuffer += string(ev.Rune())
		e.setStatusMessage(":" + e.commandBuffer)
	}
}

func (e *Editor) handleCommand() {
	parts := strings.SplitN(e.commandBuffer, " ", 2)
	command := parts[0]

	switch command {
	case "saveas":
		if len(parts) > 1 {
			newFilename := parts[1]
			if err := e.saveFileAs(newFilename); err != nil {
				e.setStatusMessage(fmt.Sprintf("Error saving as: %v", err))
			} else {
				e.SetFilename(newFilename) // Update the current filename
				e.setStatusMessage(fmt.Sprintf("File saved as %s", newFilename))
				e.isDirty = false
			}
		} else {
			e.setStatusMessage("Usage: saveas <filename>")
		}
	case "line":
		if len(parts) > 1 {
			var lineNum int
			_, err := fmt.Sscan(parts[1], &lineNum)
			if err != nil {
				e.setStatusMessage("Invalid line number")
			} else {
				if lineNum > 0 && lineNum <= len(e.lines) {
					e.cursorY = lineNum - 1
					// Keep cursor within bounds
					if e.cursorX > len(e.lines[e.cursorY]) {
						e.cursorX = len(e.lines[e.cursorY])
					}
					// Handle scrolling
					if e.cursorY < e.scrollY {
						e.scrollY = e.cursorY
					} else if e.cursorY >= e.scrollY+e.screenHeight-2 {
						e.scrollY = e.cursorY - (e.screenHeight - 3)
					}
					e.setStatusMessage(fmt.Sprintf("Moved to line %d", lineNum))
				} else {
					e.setStatusMessage("Line number out of range")
				}
			}
		} else {
			e.setStatusMessage("Usage: line <number>")
		}
	case "w":
		if err := e.saveFile(); err != nil {
			e.setStatusMessage(fmt.Sprintf("Error saving: %v", err))
		} else {
			e.setStatusMessage("File saved")
			e.isDirty = false
		}
	case "q":
		if e.isDirty {
			e.setStatusMessage("Unsaved changes! Use :q! to force quit")
		} else {
			e.quit = true
		}
	case "q!":
		e.quit = true
	case "wq":
		if err := e.saveFile(); err == nil {
			e.quit = true
		} else {
			e.setStatusMessage(fmt.Sprintf("Error saving: %v", err))
		}
	case "set number":
		e.showLineNumbers = true
		e.setStatusMessage("Line numbers enabled")
	case "set nonumber":
		e.showLineNumbers = false
		e.setStatusMessage("Line numbers disabled")
	case "rm":
		if len(parts) > 1 {
			switch parts[1] {
			case "y":
				node := e.getSelectedNode()
				if node != nil && !node.isDir {
					err := os.Remove(node.name)
					if err != nil {
						e.setStatusMessage(fmt.Sprintf("Error deleting file: %v", err))
					} else {
						e.setStatusMessage(fmt.Sprintf("Deleted %s", node.name))
						e.refreshFileTree()
					}
				}
			case "n":
				e.setStatusMessage("Delete cancelled")
			default:
				e.setStatusMessage("Invalid confirmation. Use 'rm y' or 'rm n'.")
			}
		} else {
			e.setStatusMessage("Confirm delete? (rm y/n)")
		}
	default:
		e.setStatusMessage(fmt.Sprintf("Unknown command: %s", e.commandBuffer))
	}
}

// Command implementations
func (e *Editor) saveFile() error {
	if e.filename == "" {
		return fmt.Errorf("no filename specified")
	}

	file, err := os.Create(e.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range e.lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}

func (e *Editor) saveFileAs(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range e.lines {
		if _, err := writer.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return writer.Flush()
}
