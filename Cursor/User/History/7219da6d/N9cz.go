package editor

import (
	"time"
)

// Undo/Redo functionality
type Action struct {
	Type      string
	action    string
	lines     []string
	cursorX   int
	cursorY   int
	text      string
	lineNum   int
	oldLine   string
	newLine   string
	timestamp time.Time
}

const (
	ActionInsert    = "insert"
	ActionDelete    = "delete"
	ActionJoinLines = "join_lines"
)

func (e *Editor) redo() {
	if len(e.redoStack) == 0 {
		return
	}

	// Get last redo action
	action := e.redoStack[len(e.redoStack)-1]
	e.redoStack = e.redoStack[:len(e.redoStack)-1]

	// Save current state to undo stack
	undoAction := Action{
		Type:    action.Type,
		action:  action.action,
		lines:   e.lines,
		cursorX: e.cursorX,
		cursorY: e.cursorY,
		text:    action.text,
	}
	e.undoStack = append(e.undoStack, undoAction)

	// Apply redo action
	e.lines[action.cursorY] = action.text
	e.cursorX = action.cursorX
	e.cursorY = action.cursorY
	e.isDirty = true
}

// Initialize history stacks in editor.go's NewEditor function
func (e *Editor) initHistory() {
	e.undoStack = make([]Action, 0)
	e.redoStack = make([]Action, 0)
}
