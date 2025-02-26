package editor

import (
	"strings"
)

// Cursor movement and manipulation
func (e *Editor) moveCursor(dx, dy int) {
	// Calculate new position
	newY := e.cursorY + dy
	newX := e.cursorX + dx

	// Ensure we stay within valid lines
	if newY >= 0 && newY < len(e.lines) {
		e.cursorY = newY

		// Handle scrolling
		if e.cursorY < e.scrollY {
			// Scroll up if cursor moves above visible area
			e.scrollY = e.cursorY
		} else if e.cursorY >= e.scrollY+e.screenHeight-2 {
			// Scroll down if cursor moves below visible area
			// -2 accounts for status bars
			e.scrollY = e.cursorY - (e.screenHeight - 3)
		}

		// Adjust X position based on new line length
		if e.cursorX > len(e.lines[e.cursorY]) {
			e.cursorX = len(e.lines[e.cursorY])
		}
	}

	// Ensure X position is valid
	if newX >= 0 && newX <= len(e.lines[e.cursorY]) {
		e.cursorX = newX
	}
}

func (e *Editor) insertRune(ch rune) {
	// Initialize lines if empty
	if len(e.lines) == 0 {
		e.lines = []string{""}
	}

	// Ensure cursor is within bounds
	if e.cursorY >= len(e.lines) {
		e.cursorY = len(e.lines) - 1
	}
	if e.cursorY < 0 {
		e.cursorY = 0
	}

	// Check line length limit
	if len(e.lines[e.cursorY]) >= maxLineLength {
		e.SetStatusMessage("Warning: Line length limit reached")
		return
	}

	// Insert the character
	line := e.lines[e.cursorY]
	if e.cursorX > len(line) {
		e.cursorX = len(line)
	}

	e.lines[e.cursorY] = line[:e.cursorX] + string(ch) + line[e.cursorX:]
	e.cursorX++
	e.isDirty = true
}

func (e *Editor) insertNewLine() {
	e.addUndo(Action{
		Type:    "insert",
		action:  "insert",
		lines:   e.lines,
		cursorX: e.cursorX,
		cursorY: e.cursorY,
		text:    "",
	})

	currentLine := e.lines[e.cursorY]

	// Calculate indentation of current line
	indent := ""
	for _, char := range currentLine {
		if char == ' ' || char == '\t' {
			indent += string(char)
		} else {
			break
		}
	}

	// Check for additional indentation triggers
	if strings.HasSuffix(strings.TrimSpace(currentLine), "{") {
		// Add one level of indentation after opening brace
		for i := 0; i < e.tabSize; i++ {
			indent += " "
		}
	}

	// Split the line at cursor position
	newLine := indent + currentLine[e.cursorX:]
	e.lines[e.cursorY] = currentLine[:e.cursorX]

	// Insert the new line
	e.lines = append(e.lines[:e.cursorY+1], append([]string{newLine}, e.lines[e.cursorY+1:]...)...)

	// Move cursor to the beginning of the new line (after indentation)
	e.cursorY++
	e.cursorX = len(indent)
	e.isDirty = true
}

func (e *Editor) backspace() {
	if e.cursorX > 0 || e.cursorY > 0 {
		e.addUndo(Action{
			Type:    "delete",
			action:  "delete",
			lines:   e.lines,
			cursorX: e.cursorX,
			cursorY: e.cursorY,
			text:    "",
		})
	}

	if e.cursorX > 0 {
		line := e.lines[e.cursorY]
		e.lines[e.cursorY] = line[:e.cursorX-1] + line[e.cursorX:]
		e.cursorX--
		e.isDirty = true
	} else if e.cursorY > 0 {
		// Join with previous line
		newX := len(e.lines[e.cursorY-1])
		e.lines[e.cursorY-1] += e.lines[e.cursorY]
		e.lines = append(e.lines[:e.cursorY], e.lines[e.cursorY+1:]...)
		e.cursorY--
		e.cursorX = newX
		e.isDirty = true
	}
}
