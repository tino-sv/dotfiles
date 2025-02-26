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
	// Check for file removal command
	if strings.HasPrefix(e.commandBuffer, "rm ") {
		if e.commandBuffer == "rm y" {
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
		} else if e.commandBuffer == "rm n" {
			e.setStatusMessage("Delete cancelled")
		} else {
			// First time "rm" command is entered
			e.setStatusMessage("Delete file? (y/n)")
			e.commandBuffer = "rm " // Reset to wait for y/n
		}
		return
	}

	// Handle other commands
	switch e.commandBuffer {
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
