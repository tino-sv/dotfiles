package editor

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

const VERSION = "v0.2.0.4"

// Display-related functionality
func (e *Editor) Draw() {
	// Clear screen only once per frame
	e.screen.Clear()

	// Calculate content area
	contentStartX := 0

	// Draw file tree if visible
	if e.treeVisible {
		e.drawFileTree()
		contentStartX = e.treeWidth + 1 // Add 1 for separator
	}

	// Calculate visible region based on scroll position
	startLine := e.scrollY
	endLine := startLine + e.screenHeight - 2 // Account for status bars

	if endLine > len(e.lines) {
		endLine = len(e.lines)
	}

	// Draw only visible content
	for y := startLine; y < endLine; y++ {
		screenY := y - startLine
		lineNum := y + 1

		// Draw line numbers if enabled
		if e.showLineNumbers {
			lineNumStr := fmt.Sprintf("%4d ", lineNum)
			lineNumStyle := tcell.StyleDefault.Foreground(tcell.ColorDarkGray)
			drawText(e.screen, contentStartX, screenY, lineNumStyle, lineNumStr)
		}

		// Calculate line offset based on line numbers
		xOffset := contentStartX
		if e.showLineNumbers {
			xOffset += 5
		}

		// Draw the line content with syntax highlighting
		if y < len(e.lines) {
			line := e.lines[y]
			styles := e.syntaxStyle(line)

			// Draw each character with its style
			for x, r := range line {
				if x < len(styles) {
					e.screen.SetContent(xOffset+x, screenY, r, nil, styles[x])
				}
			}
		}
	}

	// Draw completions if active
	if e.completionActive && len(e.completions) > 0 {
		e.drawCompletions()
	}

	// Draw status bars
	e.drawStatusBar()
	e.drawMessageBar()

	// Position cursor
	cursorX := e.cursorX
	if e.showLineNumbers {
		cursorX += 5
	}
	if e.treeVisible {
		cursorX += e.treeWidth + 1
	}

	// Only show cursor if it's in the visible area
	if e.cursorY >= e.scrollY && e.cursorY < e.scrollY+e.screenHeight-2 {
		e.screen.ShowCursor(cursorX, e.cursorY-e.scrollY)
	}

	// Update screen in one go
	e.screen.Show()
}

func drawText(screen tcell.Screen, x, y int, style tcell.Style, text string) {
	for _, r := range text {
		screen.SetContent(x, y, r, nil, style)
		x += runewidth.RuneWidth(r)
	}
}

func (e *Editor) setStatusMessage(msg string) {
	e.statusMessage = msg
	e.statusTimeout = time.Now().Add(3 * time.Second)
}

func (e *Editor) syntaxStyle(line string) []tcell.Style {
	styles := make([]tcell.Style, len(line))
	defaultStyle := tcell.StyleDefault

	// Define syntax styles
	keywordStyle := defaultStyle.Foreground(tcell.ColorPurple)
	typeStyle := defaultStyle.Foreground(tcell.ColorTeal)
	stringStyle := defaultStyle.Foreground(tcell.ColorGreen)
	numberStyle := defaultStyle.Foreground(tcell.ColorRed)
	commentStyle := defaultStyle.Foreground(tcell.ColorGray)
	functionStyle := defaultStyle.Foreground(tcell.ColorYellow)
	operatorStyle := defaultStyle.Foreground(tcell.ColorOrange)

	// Language keywords
	keywords := map[string]bool{
		"func": true, "return": true, "if": true, "else": true,
		"for": true, "range": true, "break": true, "continue": true,
		"switch": true, "case": true, "default": true,
		"package": true, "import": true, "type": true,
		"var": true, "const": true, "struct": true, "interface": true,
		"map": true, "chan": true, "go": true, "defer": true,
	}

	// Types
	types := map[string]bool{
		"string": true, "int": true, "bool": true, "float": true,
		"byte": true, "rune": true, "error": true, "uint": true,
		"int8": true, "int16": true, "int32": true, "int64": true,
		"uint8": true, "uint16": true, "uint32": true, "uint64": true,
	}

	// Fill with default style first
	for i := range styles {
		styles[i] = defaultStyle
	}

	inString := false
	inComment := false
	word := ""

	for i, char := range line {
		if inComment {
			styles[i] = commentStyle
			continue
		}

		if i < len(line)-1 && line[i:i+2] == "//" {
			inComment = true
			styles[i] = commentStyle
			continue
		}

		if char == '"' && (i == 0 || line[i-1] != '\\') {
			inString = !inString
			styles[i] = stringStyle
			continue
		}

		if inString {
			styles[i] = stringStyle
			continue
		}

		// Handle numbers
		if unicode.IsDigit(char) || (char == '.' && i+1 < len(line) && unicode.IsDigit(rune(line[i+1]))) {
			styles[i] = numberStyle
			continue
		}

		// Handle operators
		if strings.ContainsRune("+-*/%=<>!&|^~:;,.", char) {
			styles[i] = operatorStyle
			continue
		}

		// Handle words (keywords, types, functions)
		if unicode.IsLetter(char) || char == '_' {
			word += string(char)
		} else if word != "" {
			style := defaultStyle
			if keywords[word] {
				style = keywordStyle
			} else if types[word] {
				style = typeStyle
			} else if i < len(line) && line[i] == '(' {
				style = functionStyle
			}

			// Apply style to the whole word
			for j := i - len(word); j < i; j++ {
				styles[j] = style
			}
			word = ""
		}
	}

	return styles
}

func (e *Editor) drawStatusBar() {
	e.updateStatus()

	info := []string{
		e.statusLine,
	}

	status := strings.Join(info, " | ")
	maxWidth := e.screenWidth
	if len(status) > maxWidth {
		status = status[:maxWidth-3] + "..."
	}

	style := tcell.StyleDefault.
		Background(tcell.ColorDarkBlue).
		Foreground(tcell.ColorWhite)

	drawText(e.screen, 0, e.screenHeight-1, style, status)
}

func (e *Editor) showHelp() {
	helpText := []string{
		"Kiki's Text Editor Help",
		"---------------------",
		"",
		"Getting Started:",
		"  i       - Enter insert mode (for typing)",
		"  Esc     - Return to normal mode",
		"  :       - Enter command mode",
		"  ?       - Show this help",
		"",
		"Navigation:",
		"  h,j,k,l - Move cursor (left, down, up, right)",
		"  t       - Toggle file tree",
		"",
		"Editing:",
		"  i       - Start typing (insert mode)",
		"  Tab     - Show code completions (in insert mode)",
		"  u       - Undo",
		"  r       - Redo",
		"",
		"File Tree:",
		"  j,k     - Move up/down",
		"  Enter   - Open file/folder",
		"  n       - Create new file",
		"  D       - Delete file",
		"  r       - Rename file",
		"",
		"File Operations:",
		"  :w      - Save file",
		"  :saveas <filename> - Save file with a new name",
		"  :q      - Quit",
		"  :wq     - Save and quit",
		"  :line <number> - Go to a specific line number",
		"  :info   - Show file information",
		"  :wc     - Count lines, words, and characters",
		"  :reload - Reload the current file",
		"",
		"Settings:",
		"  :set number   - Show line numbers",
		"  :set nonumber - Hide line numbers",
		"  :set tabsize <n> - Set tab size",
		"  :set syntax on|off - Toggle syntax highlighting",
		"",
		"Search and Replace:",
		"  :find <text>  - Find text in file",
		"  :replace <old> <new> - Replace text in file",
		"",
		"Press any key to close help",
	}

	// Save current screen
	e.screen.Clear()

	// Center help text
	for i, line := range helpText {
		x := (e.screenWidth - len(line)) / 2
		if x < 0 {
			x = 0
		}
		drawText(e.screen, x, i, tcell.StyleDefault.Foreground(tcell.ColorWhite), line)
	}

	e.screen.Show()

	// Wait for keypress
	for {
		ev := e.screen.PollEvent()
		switch ev.(type) {
		case *tcell.EventKey:
			return
		}
	}
}

func (e *Editor) drawMessageBar() {
	// Clear the message bar
	for x := 0; x < e.screenWidth; x++ {
		e.screen.SetContent(x, e.screenHeight-1, ' ', nil, tcell.StyleDefault)
	}

	// Show status message if it exists
	if e.statusMessage != "" && time.Now().Before(e.statusTimeout) {
		drawText(e.screen, 0, e.screenHeight-1, tcell.StyleDefault, e.statusMessage)
		return
	}

	// Otherwise show context-sensitive key hints
	hintStyle := tcell.StyleDefault.Foreground(tcell.ColorDarkGray)
	var hints string

	switch e.mode {
	case "normal":
		if e.treeVisible {
			hints = "j/k:navigate  Enter:open  n:new  D:delete  r:rename  t:hide  ?:help"
		} else {
			hints = "i:insert  /:search  t:files  :w:save  :q:quit  ?:help"
		}
	case "insert":
		hints = "Tab:complete  Esc:normal mode"
	case "command":
		hints = "Enter:execute  Esc:cancel"
	case "search":
		hints = "Enter:find  n:next  N:prev  Esc:cancel"
	case "filename":
		hints = "Enter:create  Esc:cancel"
	case "rename":
		hints = "Enter:rename  Esc:cancel"
	}

	// Truncate if too long
	if len(hints) > e.screenWidth && e.screenWidth > 3 {
		hints = hints[:e.screenWidth-3] + "..."
	}

	drawText(e.screen, 0, e.screenHeight-1, hintStyle, hints)
}

func (e *Editor) showWelcomeScreen() {
	e.screen.Clear()

	// Title
	title := "Welcome to Kiki's Text Editor"
	drawText(e.screen, (e.screenWidth-len(title))/2, 2, tcell.StyleDefault.Foreground(tcell.ColorYellow).Bold(true), title)

	// Version
	version := VERSION
	drawText(e.screen, (e.screenWidth-len(version))/2, 4, tcell.StyleDefault.Foreground(tcell.ColorGray), version)

	// Quick start guide
	startY := 7
	quickStart := []string{
		"Quick Start Guide:",
		"",
		"  • Press 't' to toggle file tree",
		"  • Press 'i' to enter insert mode",
		"  • Press ':' to enter command mode",
		"  • Press '?' for help",
	}

	for i, line := range quickStart {
		drawText(e.screen, 10, startY+i, tcell.StyleDefault.Foreground(tcell.ColorWhite), line)
	}

	// Settings section
	settingsY := startY + len(quickStart) + 2
	settingsTitle := "Current Settings:"
	drawText(e.screen, 10, settingsY, tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true), settingsTitle)

	// Display key settings
	settings := []struct {
		name  string
		key   string
		value string
	}{
		{"Tab Size", "tabSize", e.settings["tabSize"]},
		{"Line Numbers", "showLineNumbers", e.settings["showLineNumbers"]},
		{"Syntax Highlighting", "syntaxHighlight", e.settings["syntaxHighlight"]},
		{"Auto Indent", "autoIndent", e.settings["autoIndent"]},
		{"Auto Complete", "autoComplete", e.settings["autoComplete"]},
	}

	for i, setting := range settings {
		// Format boolean settings as "On/Off" instead of "true/false"
		displayValue := setting.value
		if displayValue == "true" {
			displayValue = "On"
		} else if displayValue == "false" {
			displayValue = "Off"
		}

		settingText := fmt.Sprintf("  • %-20s: %s", setting.name, displayValue)
		drawText(e.screen, 10, settingsY+i+2, tcell.StyleDefault.Foreground(tcell.ColorWhite), settingText)
	}

	// Configuration tip
	configTip := fmt.Sprintf("Configuration file: %s", e.configFile)
	drawText(e.screen, 10, settingsY+len(settings)+4, tcell.StyleDefault.Foreground(tcell.ColorGray), configTip)

	// Footer
	footer := "Press any key to continue..."
	drawText(e.screen, (e.screenWidth-len(footer))/2, e.screenHeight-2, tcell.StyleDefault.Foreground(tcell.ColorGray), footer)

	e.screen.Show()

	// Wait for keypress
	for {
		ev := e.screen.PollEvent()
		switch ev.(type) {
		case *tcell.EventKey:
			e.isWelcomeScreen = false
			return
		case *tcell.EventResize:
			// Update screen dimensions
			e.screenWidth, e.screenHeight = e.screen.Size()
			e.showWelcomeScreen() // Redraw welcome screen after resize
			return
		}
	}
}

func (e *Editor) updateStatus() {
	status := []string{
		fmt.Sprintf("Line %d/%d", e.cursorY+1, len(e.lines)),
		fmt.Sprintf("Col %d", e.cursorX+1),
	}

	if e.filename != "" {
		status = append(status, filepath.Base(e.filename))
	}

	if e.isDirty {
		status = append(status, "[modified]")
	}

	if e.mode != "normal" {
		status = append(status, strings.ToUpper(e.mode))
	}

	if e.searchTerm != "" {
		matches := len(e.searchMatches)
		current := e.currentMatch + 1
		status = append(status, fmt.Sprintf("Search: %d/%d", current, matches))
	}

	e.statusLine = strings.Join(status, " | ")
}
