package editor

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

type Completion struct {
	Text        string
	Description string
}

func (e *Editor) getCompletions() []Completion {
	// Get the word under cursor
	if e.cursorY >= len(e.lines) {
		return nil
	}

	line := e.lines[e.cursorY]
	if e.cursorX > len(line) {
		return nil
	}

	start := e.cursorX
	for start > 0 && isIdentChar(rune(line[start-1])) {
		start--
	}
	prefix := line[start:e.cursorX]

	// Skip if prefix is too short
	if len(prefix) < 2 {
		return nil
	}

	// Get language-specific completions
	var completions []Completion
	switch e.detectLanguage() {
	case LangGo:
		completions = goCompletions
	case LangPython:
		completions = pythonCompletions
	case LangJavaScript:
		completions = jsCompletions
	case LangRust:
		completions = rustCompletions
	}

	// Filter completions based on prefix
	if prefix != "" {
		filtered := []Completion{}
		for _, c := range completions {
			if strings.HasPrefix(strings.ToLower(c.Text), strings.ToLower(prefix)) {
				filtered = append(filtered, c)
			}
		}
		return filtered
	}

	return completions
}

func (e *Editor) showCompletions() {
	completions := e.getCompletions()
	if len(completions) == 0 {
		return
	}

	// Store completions for selection
	e.completions = completions
	e.completionIndex = 0
	e.completionActive = true
}

// Apply the selected completion
func (e *Editor) applyCompletion() {
	if !e.completionActive || len(e.completions) == 0 {
		return
	}

	// Get the selected completion
	completion := e.completions[e.completionIndex]

	// Find the start of the current word
	line := e.lines[e.cursorY]
	wordStart := e.cursorX
	for wordStart > 0 && isIdentChar(rune(line[wordStart-1])) {
		wordStart--
	}

	// Replace the current word with the completion
	newLine := line[:wordStart] + completion.Text + line[e.cursorX:]
	e.addUndo(Action{
		Type:    "insert",
		action:  "insert",
		lines:   e.lines,
		cursorX: e.cursorX,
		cursorY: e.cursorY,
		text:    completion.Text,
	})
	e.lines[e.cursorY] = newLine
	e.cursorX = wordStart + len(completion.Text)
	e.isDirty = true

	// Clear completion state
	e.completionActive = false
}

// Navigate through completions
func (e *Editor) nextCompletion() {
	if e.completionActive && len(e.completions) > 0 {
		e.completionIndex = (e.completionIndex + 1) % len(e.completions)
	}
}

func (e *Editor) prevCompletion() {
	if e.completionActive && len(e.completions) > 0 {
		e.completionIndex--
		if e.completionIndex < 0 {
			e.completionIndex = len(e.completions) - 1
		}
	}
}

func (e *Editor) drawCompletions() {
	// Calculate position for completion popup
	popupX := e.cursorX
	if e.showLineNumbers {
		popupX += 5
	}
	if e.treeVisible {
		popupX += e.treeWidth + 1
	}

	popupY := e.cursorY - e.scrollY + 1 // Show below cursor

	// Ensure popup fits on screen
	if popupY >= e.screenHeight-3 {
		popupY = e.cursorY - e.scrollY - len(e.completions) - 1 // Show above cursor
	}

	// Calculate max width needed
	maxWidth := 0
	for _, c := range e.completions {
		width := len(c.Text) + len(c.Description) + 3 // +3 for spacing
		if width > maxWidth {
			maxWidth = width
		}
	}

	// Draw popup background
	popupStyle := tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite)
	selectedStyle := tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite)

	for i, comp := range e.completions {
		// Limit number of displayed completions
		if i >= 10 {
			break
		}

		// Choose style based on selection
		style := popupStyle
		if i == e.completionIndex {
			style = selectedStyle
		}

		// Draw background
		for x := 0; x < maxWidth; x++ {
			e.screen.SetContent(popupX+x, popupY+i, ' ', nil, style)
		}

		// Draw completion text
		drawText(e.screen, popupX, popupY+i, style, comp.Text)

		// Draw description
		descStyle := style.Foreground(tcell.ColorLightGray)
		drawText(e.screen, popupX+len(comp.Text)+2, popupY+i, descStyle, comp.Description)
	}
}

func isIdentChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// Language-specific completions
var goCompletions = []Completion{
	// Go keywords
	{Text: "func", Description: "function declaration"},
	{Text: "type", Description: "type declaration"},
	{Text: "struct", Description: "struct declaration"},
	{Text: "interface", Description: "interface declaration"},
	{Text: "map", Description: "map type"},
	{Text: "chan", Description: "channel type"},
	{Text: "if", Description: "if statement"},
	{Text: "else", Description: "else statement"},
	{Text: "for", Description: "for loop"},
	{Text: "range", Description: "range clause"},
	{Text: "return", Description: "return statement"},
	{Text: "package", Description: "package declaration"},
	{Text: "import", Description: "import declaration"},
	{Text: "var", Description: "variable declaration"},
	{Text: "const", Description: "constant declaration"},

	// Common types
	{Text: "string", Description: "string type"},
	{Text: "int", Description: "integer type"},
	{Text: "bool", Description: "boolean type"},
	{Text: "error", Description: "error type"},
	{Text: "float64", Description: "64-bit float"},

	// Common snippets
	{Text: "fmt.Println()", Description: "print line"},
	{Text: "fmt.Printf()", Description: "print formatted"},
	{Text: "make()", Description: "make builtin"},
	{Text: "append()", Description: "append builtin"},
	{Text: "len()", Description: "length builtin"},

	// Additional Go completions
	{Text: "defer", Description: "defer execution"},
	{Text: "go", Description: "start goroutine"},
	{Text: "select", Description: "select statement"},
	{Text: "switch", Description: "switch statement"},
	{Text: "case", Description: "case clause"},
	{Text: "default", Description: "default clause"},
	{Text: "break", Description: "break statement"},
	{Text: "continue", Description: "continue statement"},
	{Text: "fallthrough", Description: "fallthrough statement"},
	{Text: "goto", Description: "goto statement"},
}

var pythonCompletions = []Completion{
	// Python keywords
	{Text: "def", Description: "function definition"},
	{Text: "class", Description: "class definition"},
	{Text: "if", Description: "if statement"},
	{Text: "elif", Description: "else if statement"},
	{Text: "else", Description: "else statement"},
	{Text: "for", Description: "for loop"},
	{Text: "while", Description: "while loop"},
	{Text: "try", Description: "try block"},
	{Text: "except", Description: "except block"},
	{Text: "finally", Description: "finally block"},
	{Text: "import", Description: "import statement"},
	{Text: "from", Description: "from import"},
	{Text: "return", Description: "return statement"},

	// Common functions
	{Text: "print()", Description: "print function"},
	{Text: "len()", Description: "length function"},
	{Text: "range()", Description: "range function"},
	{Text: "list()", Description: "list constructor"},
	{Text: "dict()", Description: "dictionary constructor"},
	{Text: "set()", Description: "set constructor"},

	// Additional Python completions
	{Text: "with", Description: "with statement"},
	{Text: "as", Description: "as clause"},
	{Text: "lambda", Description: "lambda expression"},
	{Text: "yield", Description: "yield statement"},
	{Text: "global", Description: "global statement"},
	{Text: "nonlocal", Description: "nonlocal statement"},
	{Text: "assert", Description: "assert statement"},
	{Text: "raise", Description: "raise exception"},
	{Text: "pass", Description: "pass statement"},
	{Text: "break", Description: "break statement"},
	{Text: "continue", Description: "continue statement"},
}

var jsCompletions = []Completion{
	// JavaScript keywords
	{Text: "function", Description: "function declaration"},
	{Text: "const", Description: "constant declaration"},
	{Text: "let", Description: "block-scoped variable"},
	{Text: "var", Description: "function-scoped variable"},
	{Text: "if", Description: "if statement"},
	{Text: "else", Description: "else statement"},
	{Text: "for", Description: "for loop"},
	{Text: "while", Description: "while loop"},
	{Text: "class", Description: "class declaration"},
	{Text: "return", Description: "return statement"},
	{Text: "import", Description: "import statement"},
	{Text: "export", Description: "export statement"},

	// Common methods
	{Text: "console.log()", Description: "console log"},
	{Text: "console.error()", Description: "console error"},
	{Text: "Array.from()", Description: "array from iterable"},
	{Text: "Object.keys()", Description: "object keys"},
	{Text: "Promise", Description: "Promise constructor"},
	{Text: "async", Description: "async function"},
	{Text: "await", Description: "await expression"},

	// Additional JS completions
	{Text: "document", Description: "DOM document"},
	{Text: "window", Description: "browser window"},
	{Text: "setTimeout()", Description: "set timeout"},
	{Text: "setInterval()", Description: "set interval"},
	{Text: "addEventListener()", Description: "add event listener"},
	{Text: "querySelector()", Description: "query selector"},
	{Text: "querySelectorAll()", Description: "query selector all"},
	{Text: "getElementById()", Description: "get element by ID"},
	{Text: "createElement()", Description: "create element"},
}

var rustCompletions = []Completion{
	// Rust keywords
	{Text: "fn", Description: "function declaration"},
	{Text: "struct", Description: "struct declaration"},
	{Text: "enum", Description: "enum declaration"},
	{Text: "impl", Description: "implementation block"},
	{Text: "trait", Description: "trait declaration"},
	{Text: "let", Description: "variable binding"},
	{Text: "mut", Description: "mutable binding"},
	{Text: "if", Description: "if expression"},
	{Text: "else", Description: "else expression"},
	{Text: "match", Description: "match expression"},
	{Text: "loop", Description: "loop expression"},
	{Text: "while", Description: "while loop"},
	{Text: "for", Description: "for loop"},

	// Common macros
	{Text: "println!()", Description: "print line macro"},
	{Text: "vec![]", Description: "vector macro"},
	{Text: "Some()", Description: "Some variant"},
	{Text: "None", Description: "None variant"},
	{Text: "Ok()", Description: "Ok variant"},
	{Text: "Err()", Description: "Err variant"},

	// Additional Rust completions
	{Text: "pub", Description: "public visibility"},
	{Text: "use", Description: "use declaration"},
	{Text: "mod", Description: "module declaration"},
	{Text: "crate", Description: "crate reference"},
	{Text: "self", Description: "self reference"},
	{Text: "super", Description: "parent module"},
	{Text: "return", Description: "return expression"},
	{Text: "break", Description: "break expression"},
	{Text: "continue", Description: "continue expression"},
	{Text: "unsafe", Description: "unsafe block"},
	{Text: "extern", Description: "external block"},
}
