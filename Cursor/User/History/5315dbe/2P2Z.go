// Main editor package
package editor

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Add buffer size limit
const maxLineLength = 10000

type Editor struct {
	screen           tcell.Screen
	lines            []string
	cursorX, cursorY int
	mode             string
	filename         string
	statusMessage    string
	statusTimeout    time.Time
	isDirty          bool
	tabSize          int
	searchTerm       string
	searchMatches    []struct{ y, x int }
	currentMatch     int
	undoStack        []Action
	redoStack        []Action
	commandBuffer    string
	showLineNumbers  bool
	quit             bool
	treeVisible      bool
	treeWidth        int
	currentPath      string
	fileTree         *FileNode
	treeSelectedLine int
	screenWidth      int
	screenHeight     int
	newFileDir       string
	isWelcomeScreen  bool
	confirmAction    func()
	scrollY          int // Vertical scroll position
	syntaxHighlight  bool

	// Auto-completion fields
	completions      []Completion
	completionIndex  int
	completionActive bool

	// User settings
	settings   map[string]string
	configFile string
	wordWrap   bool
}

func (e *Editor) SetFilename(name string) {
	e.filename = name
}

func NewEditor() (*Editor, error) {
	// Initialize screen
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}

	// Enable mouse support
	screen.EnableMouse()

	// Get screen dimensions
	width, height := screen.Size()

	// Create editor instance
	ed := &Editor{
		screen:          screen,
		lines:           []string{""},
		mode:            "normal",
		tabSize:         4,
		showLineNumbers: true,
		treeVisible:     true,
		treeWidth:       30,
		screenWidth:     width,
		screenHeight:    height,
		undoStack:       make([]Action, 0),
		redoStack:       make([]Action, 0),
		isWelcomeScreen: true,
		wordWrap:        false,
	}

	ed.initFileTree()
	ed.SetStatusMessage("Welcome! Press '?' for help, 'i' for insert mode, ':' for commands")

	// Show welcome screen
	ed.showWelcomeScreen()

	ed.initHistory()

	return ed, nil
}

func (e *Editor) Run() {
	// Basic nil checks
	if e == nil || e.screen == nil {
		log.Fatal("Editor or screen not properly initialized")
	}

	// Defer screen cleanup
	defer e.screen.Fini()

	for {
		e.updateScreenSize()
		e.Draw()

		// Handle events
		ev := e.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlC {
				return
			}
			e.handleInput(ev)
		case *tcell.EventMouse:
			e.handleMouseEvent(ev)
		case *tcell.EventResize:
			e.screen.Sync()
			e.updateScreenSize()
		}

		if e.quit {
			return
		}
	}
}

func (e *Editor) updateScreenSize() {
	e.screenWidth, e.screenHeight = e.screen.Size()
}

func (e *Editor) SetStatusMessage(msg string) {
	e.statusMessage = msg
	e.statusTimeout = time.Now().Add(3 * time.Second)
}

func (e *Editor) deleteChar() {
	if e.cursorY >= len(e.lines) || e.cursorX <= 0 || e.cursorX > len(e.lines[e.cursorY]) {
		return
	}

	line := e.lines[e.cursorY]
	e.lines[e.cursorY] = line[:e.cursorX-1] + line[e.cursorX:]
	e.cursorX--
	e.isDirty = true

	e.undoStack = append(e.undoStack, Action{
		Type:    ActionDelete,
		lineNum: e.cursorY,
		oldLine: line,
		newLine: e.lines[e.cursorY],
		cursorX: e.cursorX,
		cursorY: e.cursorY,
		text:    string(line[e.cursorX]),
	})
	e.redoStack = nil
}

func (e *Editor) joinLines() {
	if e.cursorY > 0 {
		currentLine := e.lines[e.cursorY]
		prevLine := e.lines[e.cursorY-1]
		e.cursorX = len(prevLine)
		e.lines[e.cursorY-1] = prevLine + currentLine
		e.lines = append(e.lines[:e.cursorY], e.lines[e.cursorY+1:]...)
		e.cursorY--
		e.isDirty = true

		// Record action for undo
		e.undoStack = append(e.undoStack, Action{
			Type:    ActionJoinLines,
			lineNum: e.cursorY,
			oldLine: currentLine,
			newLine: prevLine + currentLine,
			cursorX: e.cursorX,
			cursorY: e.cursorY,
			text:    "\n",
		})
		e.redoStack = nil
	}
}

func (e *Editor) getFileType() string {
	if e.filename == "" {
		return "New File"
	}
	ext := filepath.Ext(e.filename)
	if ext == "" {
		return "Text"
	}
	return strings.TrimPrefix(ext, ".")
}

func (e *Editor) getFileSize() int64 {
	if e.filename == "" {
		return 0
	}

	file, err := os.Stat(e.filename)
	if err != nil {
		return 0
	}
	return file.Size()
}

func (e *Editor) insertRune(ch rune) {
	if len(e.lines[e.cursorY]) >= maxLineLength {
		e.SetStatusMessage("Warning: Line length limit reached")
		return
	}
	// ... rest of insertRune implementation
}
