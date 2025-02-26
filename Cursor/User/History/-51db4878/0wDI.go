package editor

import (
	"regexp"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Syntax highlighting patterns
var (
	// Go patterns
	goKeywords = []string{
		"break", "default", "func", "interface", "select", "case", "defer",
		"go", "map", "struct", "chan", "else", "goto", "package", "switch",
		"const", "fallthrough", "if", "range", "type", "continue", "for",
		"import", "return", "var",
	}
	goTypes = []string{"string", "int", "bool", "byte", "rune", "float", "error"}

	// Python patterns
	pythonKeywords = []string{
		"and", "as", "assert", "break", "class", "continue", "def", "del",
		"elif", "else", "except", "False", "finally", "for", "from", "global",
		"if", "import", "in", "is", "lambda", "None", "nonlocal", "not", "or",
		"pass", "raise", "return", "True", "try", "while", "with", "yield",
	}

	// JavaScript patterns
	jsKeywords = []string{
		"break", "case", "catch", "class", "const", "continue", "debugger",
		"default", "delete", "do", "else", "export", "extends", "finally",
		"for", "function", "if", "import", "in", "instanceof", "new", "return",
		"super", "switch", "this", "throw", "try", "typeof", "var", "void",
		"while", "with", "yield", "let", "static", "async", "await",
	}

	// Rust patterns
	rustKeywords = []string{
		"as", "break", "const", "continue", "crate", "else", "enum", "extern",
		"false", "fn", "for", "if", "impl", "in", "let", "loop", "match", "mod",
		"move", "mut", "pub", "ref", "return", "self", "Self", "static", "struct",
		"super", "trait", "true", "type", "unsafe", "use", "where", "while",
	}

	// Common patterns
	numberPattern  = regexp.MustCompile(`\b\d+(\.\d+)?\b`)
	stringPattern  = regexp.MustCompile(`"[^"]*"`)
	commentPattern = regexp.MustCompile(`//.*$|/\*[\s\S]*?\*/|#.*$`)
)

// Highlight styles
var (
	keywordStyle = tcell.StyleDefault.Foreground(tcell.ColorAqua)
	typeStyle    = tcell.StyleDefault.Foreground(tcell.ColorGreen)
	stringStyle  = tcell.StyleDefault.Foreground(tcell.ColorYellow)
	numberStyle  = tcell.StyleDefault.Foreground(tcell.ColorPurple)
	commentStyle = tcell.StyleDefault.Foreground(tcell.ColorGray)
	defaultStyle = tcell.StyleDefault
)

type SyntaxCache struct {
	line      string
	styles    []tcell.Style
	timestamp time.Time
}

func (e *Editor) highlightSyntax(line string) []tcell.Style {
	// Check cache first
	if cache, ok := e.syntaxCache[line]; ok {
		return cache
	}

	// Calculate styles
	styles := make([]tcell.Style, len(line))

	// Fill with default style
	for i := range styles {
		styles[i] = defaultStyle
	}

	// Language-specific highlighting
	switch e.detectLanguage() {
	case LangGo:
		highlightGo(line, styles)
	case LangPython:
		highlightPython(line, styles)
	case LangJavaScript:
		highlightJavaScript(line, styles)
	case LangRust:
		highlightRust(line, styles)
	}

	// Common patterns (numbers, strings, comments)
	highlightPattern(line, styles, numberPattern, numberStyle)
	highlightPattern(line, styles, stringPattern, stringStyle)
	highlightPattern(line, styles, commentPattern, commentStyle)

	// Cache the result
	e.syntaxCache[line] = styles

	return styles
}

func highlightGo(line string, styles []tcell.Style) {
	highlightKeywords(line, styles, goKeywords, keywordStyle)
	highlightKeywords(line, styles, goTypes, typeStyle)
}

func highlightPython(line string, styles []tcell.Style) {
	highlightKeywords(line, styles, pythonKeywords, keywordStyle)
}

func highlightJavaScript(line string, styles []tcell.Style) {
	highlightKeywords(line, styles, jsKeywords, keywordStyle)
}

func highlightRust(line string, styles []tcell.Style) {
	highlightKeywords(line, styles, rustKeywords, keywordStyle)
}

func highlightKeywords(line string, styles []tcell.Style, keywords []string, style tcell.Style) {
	for _, keyword := range keywords {
		pattern := regexp.MustCompile(`\b` + keyword + `\b`)
		highlightPattern(line, styles, pattern, style)
	}
}

func highlightPattern(line string, styles []tcell.Style, pattern *regexp.Regexp, style tcell.Style) {
	matches := pattern.FindAllStringIndex(line, -1)
	for _, match := range matches {
		start, end := match[0], match[1]
		for i := start; i < end && i < len(styles); i++ {
			styles[i] = style
		}
	}
}
