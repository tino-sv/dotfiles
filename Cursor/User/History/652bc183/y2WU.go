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
	case "info":
		if e.filename != "" {
			fileInfo, err := os.Stat(e.filename)
			if err == nil {
				infoMsg := fmt.Sprintf("File: %s | Size: %d bytes | Modified: %s",
					e.filename,
					fileInfo.Size(),
					fileInfo.ModTime().Format("2006-01-02 15:04:05"))
				e.setStatusMessage(infoMsg)
			} else {
				e.setStatusMessage("Could not retrieve file information")
			}
		} else {
			e.setStatusMessage("No file currently open")
		}
	case "wc":
		wordCount := 0
		lineCount := len(e.lines)
		charCount := 0

		for _, line := range e.lines {
			words := strings.Fields(line)
			wordCount += len(words)
			charCount += len(line)
		}

		infoMsg := fmt.Sprintf("Lines: %d | Words: %d | Characters: %d",
			lineCount, wordCount, charCount)
		e.setStatusMessage(infoMsg)
	case "reload":
		if e.filename != "" {
			err := e.LoadFile(e.filename)
			if err != nil {
				e.setStatusMessage(fmt.Sprintf("Error reloading file: %v", err))
			} else {
				e.setStatusMessage(fmt.Sprintf("Reloaded: %s", e.filename))
				e.isDirty = false
			}
		} else {
			e.setStatusMessage("No file to reload")
		}
	case "set":
		if len(parts) > 1 {
			settingParts := strings.SplitN(parts[1], " ", 2)
			settingName := settingParts[0]

			switch settingName {
			case "tabsize":
				if len(settingParts) > 1 {
					tabSize, err := fmt.Sscan(settingParts[1], &tabSize)
					if err == nil && tabSize > 0 {
						e.tabSize = tabSize
						e.setStatusMessage(fmt.Sprintf("Tab size set to %d", tabSize))
					} else {
						e.setStatusMessage("Invalid tab size")
					}
				} else {
					e.setStatusMessage("Usage: set tabsize <number>")
				}
			case "syntax":
				if len(settingParts) > 1 {
					if settingParts[1] == "on" {
						e.syntaxHighlight = true
						e.setStatusMessage("Syntax highlighting enabled")
					} else if settingParts[1] == "off" {
						e.syntaxHighlight = false
						e.setStatusMessage("Syntax highlighting disabled")
					} else {
						e.setStatusMessage("Usage: set syntax on|off")
					}
				} else {
					e.setStatusMessage("Usage: set syntax on|off")
				}
			}
		}
	case "find":
		if len(parts) > 1 {
			e.searchTerm = parts[1]
			e.findMatches()
			if len(e.searchMatches) > 0 {
				e.setStatusMessage(fmt.Sprintf("Found %d matches", len(e.searchMatches)))
			} else {
				e.setStatusMessage("No matches found")
			}
		} else {
			e.setStatusMessage("Usage: find <search term>")
		}
	case "replace":
		if len(parts) > 1 {
			replaceParts := strings.SplitN(parts[1], " ", 2)
			if len(replaceParts) == 2 {
				oldText := replaceParts[0]
				newText := replaceParts[1]
				count := 0

				// Save state for undo
				for i, line := range e.lines {
					if strings.Contains(line, oldText) {
						e.addUndo(line)
						e.lines[i] = strings.ReplaceAll(line, oldText, newText)
						count += strings.Count(line, oldText)
						e.isDirty = true
					}
				}

				if count > 0 {
					e.setStatusMessage(fmt.Sprintf("Replaced %d occurrences", count))
				} else {
					e.setStatusMessage("No matches found")
				}
			} else {
				e.setStatusMessage("Usage: replace <old> <new>")
			}
		} else {
			e.setStatusMessage("Usage: replace <old> <new>")
		}
	case "help":
		e.showHelp()
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
