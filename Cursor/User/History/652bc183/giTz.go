package editor

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Command mode functionality

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
			lineNum, err := strconv.Atoi(parts[1])
			if err == nil && lineNum > 0 && lineNum <= len(e.lines) {
				e.cursorY = lineNum - 1
				e.SetStatusMessage(fmt.Sprintf("Jumped to line %d", lineNum))
			} else {
				e.SetStatusMessage("Invalid line number")
			}
		} else {
			e.SetStatusMessage("Usage: line <number>")
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
	case "set":
		if len(parts) > 1 {
			settingParts := strings.SplitN(parts[1], " ", 2)
			settingName := settingParts[0]

			switch settingName {
			case "tabsize":
				if len(settingParts) > 1 {
					var newTabSize int
					_, err := fmt.Sscan(settingParts[1], &newTabSize)
					if err == nil && newTabSize > 0 {
						e.tabSize = newTabSize
						e.setSetting("tabSize", settingParts[1])
						e.setStatusMessage(fmt.Sprintf("Tab size set to %d", newTabSize))
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
						e.setSetting("syntaxHighlight", "true")
						e.setStatusMessage("Syntax highlighting enabled")
					} else if settingParts[1] == "off" {
						e.syntaxHighlight = false
						e.setSetting("syntaxHighlight", "false")
						e.setStatusMessage("Syntax highlighting disabled")
					} else {
						e.setStatusMessage("Usage: set syntax on|off")
					}
				} else {
					e.setStatusMessage("Usage: set syntax on|off")
				}
			case "number":
				e.showLineNumbers = !e.showLineNumbers
				e.setStatusMessage(fmt.Sprintf("Line numbers %s", map[bool]string{true: "enabled", false: "disabled"}[e.showLineNumbers]))
			case "nonumber":
				e.showLineNumbers = false
				e.setSetting("showLineNumbers", "false")
				e.setStatusMessage("Line numbers disabled")
			case "autoindent":
				if len(settingParts) > 1 {
					if settingParts[1] == "on" {
						e.setSetting("autoIndent", "true")
						e.setStatusMessage("Auto-indent enabled")
					} else if settingParts[1] == "off" {
						e.setSetting("autoIndent", "false")
						e.setStatusMessage("Auto-indent disabled")
					} else {
						e.setStatusMessage("Usage: set autoindent on|off")
					}
				} else {
					e.setStatusMessage("Usage: set autoindent on|off")
				}
			case "autocomplete":
				if len(settingParts) > 1 {
					if settingParts[1] == "on" {
						e.setSetting("autoComplete", "true")
						e.setStatusMessage("Auto-complete enabled")
					} else if settingParts[1] == "off" {
						e.setSetting("autoComplete", "false")
						e.setStatusMessage("Auto-complete disabled")
					} else {
						e.setStatusMessage("Usage: set autocomplete on|off")
					}
				} else {
					e.setStatusMessage("Usage: set autocomplete on|off")
				}
			case "wrap":
				e.wordWrap = !e.wordWrap
				e.setStatusMessage(fmt.Sprintf("Word wrap %s", map[bool]string{true: "enabled", false: "disabled"}[e.wordWrap]))
			default:
				e.setStatusMessage(fmt.Sprintf("Unknown setting: %s", settingName))
			}
		} else {
			// Show all settings
			settingsStr := "Current settings:\n"
			for k, v := range e.settings {
				settingsStr += fmt.Sprintf("%s = %s\n", k, v)
			}
			e.setStatusMessage(settingsStr)
		}
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
		info := fmt.Sprintf("File: %s\nLines: %d\nSize: %d bytes\nType: %s",
			e.filename,
			len(e.lines),
			e.getFileSize(),
			e.getFileType())
		e.SetStatusMessage(info)
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
	case "find":
		if len(parts) > 1 {
			e.searchTerm = parts[1]
			e.findNext()
		} else {
			e.setStatusMessage("Usage: find <text>")
		}
	case "replace":
		if len(parts) > 2 {
			oldText := parts[1]
			newText := parts[2]
			count := e.replaceAll(oldText, newText)
			e.setStatusMessage(fmt.Sprintf("Replaced %d occurrences", count))
		} else {
			e.setStatusMessage("Usage: replace <old> <new>")
		}
	case "help":
		e.showHelp()
	case "delete":
		if len(parts) > 1 {
			filename := parts[1]
			err := os.Remove(filename)
			if err != nil {
				e.setStatusMessage(fmt.Sprintf("Error deleting file: %v", err))
			} else {
				e.setStatusMessage("File deleted")
			}
		} else {
			e.setStatusMessage("Usage: delete <filename>")
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
