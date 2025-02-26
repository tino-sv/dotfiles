package editor

// Undo/Redo functionality
type Action struct {
	Type    string
	lineNum int
	oldLine string
	newLine string
	cursorX int
	cursorY int
	text    string
}

const (
	ActionInsert    = "insert"
	ActionDelete    = "delete"
	ActionJoinLines = "join_lines"
)

func (e *Editor) addUndo(oldLine string) {
	action := Action{
		lineNum: e.cursorY,
		oldLine: oldLine,
		newLine: e.lines[e.cursorY],
		cursorX: e.cursorX,
		cursorY: e.cursorY,
	}
	e.undoStack = append(e.undoStack, action)
	// Clear redo stack when new action is performed
	e.redoStack = nil
}

func (e *Editor) undo() {
	if len(e.undoStack) == 0 {
		return
	}

	// Get last action
	action := e.undoStack[len(e.undoStack)-1]
	e.undoStack = e.undoStack[:len(e.undoStack)-1]

	// Save current state to redo stack
	redoAction := Action{
		lineNum: action.lineNum,
		oldLine: e.lines[action.lineNum],
		newLine: action.oldLine,
		cursorX: e.cursorX,
		cursorY: e.cursorY,
	}
	e.redoStack = append(e.redoStack, redoAction)

	// Restore previous state
	e.lines[action.lineNum] = action.oldLine
	e.cursorX = action.cursorX
	e.cursorY = action.cursorY
	e.isDirty = true
}

func (e *Editor) redo() {
	if len(e.redoStack) == 0 {
		return
	}

	// Get last redo action
	action := e.redoStack[len(e.redoStack)-1]
	e.redoStack = e.redoStack[:len(e.redoStack)-1]

	// Save current state to undo stack
	undoAction := Action{
		lineNum: action.lineNum,
		oldLine: e.lines[action.lineNum],
		newLine: action.newLine,
		cursorX: e.cursorX,
		cursorY: e.cursorY,
	}
	e.undoStack = append(e.undoStack, undoAction)

	// Apply redo action
	e.lines[action.lineNum] = action.newLine
	e.cursorX = action.cursorX
	e.cursorY = action.cursorY
	e.isDirty = true
}

// Initialize history stacks in editor.go's NewEditor function
func (e *Editor) initHistory() {
	e.undoStack = make([]Action, 0)
	e.redoStack = make([]Action, 0)
}
