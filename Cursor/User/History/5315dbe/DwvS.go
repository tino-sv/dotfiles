// Main editor package
package editor

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Add buffer size limit
const (
	maxLineLength = 10000
	maxUndoStack  = 1000
	maxFileSize   = 50 * 1024 * 1024 // 50MB
)

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

	// Auto-completion fields
	completions      []Completion
	completionIndex  int
	completionActive bool

	// User settings
	settings   map[string]string
	configFile string

	lineCache       map[int]string // Cache for long lines
	syntaxCache     map[string][]tcell.Style
	isLargeFile     bool
	showLineNumbers bool
	syntaxHighlight bool
	debugMode       bool
	wordWrap        bool
	statusLine      string
	searchIndex     SearchIndex
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
		debugMode:       false,
		syntaxHighlight: true,
		wordWrap:        false,
		statusLine:      "",
		searchIndex: SearchIndex{
			positions: make(map[string][]Position),
			dirty:     true,
		},
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
		lines:   e.lines,
		cursorX: e.cursorX,
		cursorY: e.cursorY,
		action:  "delete",
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
			lines:   e.lines,
			cursorX: e.cursorX,
			cursorY: e.cursorY,
			action:  "join",
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

func (e *Editor) addUndo(action Action) {
	e.undoStack = append(e.undoStack, action)
	// Clear redo stack when new change is made
	e.redoStack = nil
}

func (e *Editor) undo() {
	if len(e.undoStack) == 0 {
		return
	}

	// Save current state to redo stack
	currentState := Action{
		lines:   append([]string{}, e.lines...),
		cursorX: e.cursorX,
		cursorY: e.cursorY,
		action:  "redo",
	}
	e.redoStack = append(e.redoStack, currentState)

	// Restore previous state
	action := e.undoStack[len(e.undoStack)-1]
	e.undoStack = e.undoStack[:len(e.undoStack)-1]
	e.lines = append([]string{}, action.lines...)
	e.cursorX = action.cursorX
	e.cursorY = action.cursorY
}

func (e *Editor) SaveFile() error {
	if e.filename == "" {
		return fmt.Errorf("no filename specified")
	}

	// Create backup before saving
	if err := e.createBackup(); err != nil {
		return fmt.Errorf("backup failed: %v", err)
	}

	// Write to temporary file first
	tempFile := e.filename + ".tmp"
	if err := e.writeToFile(tempFile); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("write failed: %v", err)
	}

	// Rename temporary file to actual file
	if err := os.Rename(tempFile, e.filename); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("rename failed: %v", err)
	}

	e.isDirty = false
	return nil
}

func (e *Editor) createBackup() error {
	if e.filename == "" {
		return fmt.Errorf("no filename specified")
	}

	backupFile := e.filename + ".bak"
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return nil
	}

	return fmt.Errorf("backup file already exists")
}

func (e *Editor) writeToFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, line := range e.lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return err
		}
	}

	return nil
}

func (e *Editor) LoadFile(filename string) error {
	info, err := os.Stat(filename)
	if err != nil {
		return err
	}

	// Check file size
	if info.Size() > maxFileSize {
		e.isLargeFile = true
		return e.loadLargeFile(filename)
	}

	// Normal file loading...
	return e.loadNormalFile(filename)
}

func (e *Editor) loadLargeFile(filename string) error {
	// Load file in chunks
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	e.lines = make([]string, 0)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, maxLineLength), maxLineLength)

	for scanner.Scan() {
		e.lines = append(e.lines, scanner.Text())
		if len(e.lines) > 1000 {
			// Only load first 1000 lines initially
			break
		}
	}

	e.SetStatusMessage("Large file: Only first 1000 lines loaded")
	return nil
}

func (e *Editor) loadNormalFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	e.lines = make([]string, 0)
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, maxLineLength), maxLineLength)

	for scanner.Scan() {
		e.lines = append(e.lines, scanner.Text())
	}

	return nil
}

func (e *Editor) replaceAll(old, new string) int {
	count := 0
	for i, line := range e.lines {
		if strings.Contains(line, old) {
			e.lines[i] = strings.ReplaceAll(line, old, new)
			count += strings.Count(line, old)
			e.isDirty = true
		}
	}
	return count
}

func (e *Editor) EnableDebugMode() {
	debugPath := filepath.Join(os.Getenv("HOME"), ".kiki-editor")
	if err := os.MkdirAll(debugPath, 0755); err != nil {
		return
	}

	logFile := filepath.Join(debugPath, "debug.log")
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}

	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	e.debugMode = true
}
