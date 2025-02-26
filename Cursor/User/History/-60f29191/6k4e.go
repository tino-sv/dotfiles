package editor

import (
	"path/filepath"
	"strings"
)

// Completion candidate
type Completion struct {
	Text        string
	Description string
}

// Get completion candidates based on current context
func (e *Editor) getCompletions() []Completion {
	// Get current line and word being typed
	line := e.lines[e.cursorY]
	wordStart := e.cursorX

	// Find the start of the current word
	for wordStart > 0 && isWordChar(rune(line[wordStart-1])) {
		wordStart--
	}

	// Extract the prefix (partial word)
	prefix := ""
	if wordStart < e.cursorX && wordStart < len(line) {
		prefix = line[wordStart:e.cursorX]
	}

	// Skip if prefix is too short
	if len(prefix) < 2 {
		return nil
	}

	// Get language-specific completions
	completions := e.getLanguageCompletions(prefix)

	// Add file-specific completions
	if e.filename != "" {
		ext := filepath.Ext(e.filename)
		completions = append(completions, e.getFileTypeCompletions(prefix, ext)...)
	}

	return completions
}

// Check if a character is part of a word
func isWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// Get language-specific completions
func (e *Editor) getLanguageCompletions(prefix string) []Completion {
	// Common programming keywords
	keywords := []Completion{
		{"func", "Function declaration"},
		{"return", "Return statement"},
		{"if", "Conditional statement"},
		{"else", "Else clause"},
		{"for", "Loop statement"},
		{"while", "While loop"},
		{"switch", "Switch statement"},
		{"case", "Case clause"},
		{"break", "Break statement"},
		{"continue", "Continue statement"},
		{"struct", "Structure definition"},
		{"interface", "Interface definition"},
		{"class", "Class definition"},
		{"import", "Import statement"},
		{"package", "Package declaration"},
		{"const", "Constant declaration"},
		{"var", "Variable declaration"},
		{"type", "Type definition"},
		{"map", "Map data structure"},
		{"slice", "Slice data structure"},
		{"array", "Array data structure"},
		{"string", "String type"},
		{"int", "Integer type"},
		{"float", "Float type"},
		{"bool", "Boolean type"},
		{"true", "Boolean true"},
		{"false", "Boolean false"},
		{"nil", "Nil value"},
	}

	// Filter by prefix
	var filtered []Completion
	for _, k := range keywords {
		if strings.HasPrefix(k.Text, prefix) {
			filtered = append(filtered, k)
		}
	}

	return filtered
}

// Get file-type specific completions
func (e *Editor) getFileTypeCompletions(prefix string, ext string) []Completion {
	var completions []Completion

	switch ext {
	case ".go":
		goKeywords := []Completion{
			{"defer", "Defer execution"},
			{"go", "Start goroutine"},
			{"chan", "Channel type"},
			{"select", "Select statement"},
			{"make", "Allocate and initialize"},
			{"new", "Allocate memory"},
			{"append", "Append to slice"},
			{"len", "Length function"},
			{"cap", "Capacity function"},
			{"panic", "Panic function"},
			{"recover", "Recover function"},
		}

		for _, k := range goKeywords {
			if strings.HasPrefix(k.Text, prefix) {
				completions = append(completions, k)
			}
		}
	case ".js", ".ts":
		jsKeywords := []Completion{
			{"function", "Function declaration"},
			{"const", "Constant declaration"},
			{"let", "Block-scoped variable"},
			{"var", "Variable declaration"},
			{"console.log", "Print to console"},
			{"document", "DOM document"},
			{"window", "Browser window"},
			{"setTimeout", "Set timeout"},
			{"setInterval", "Set interval"},
			{"Promise", "Promise object"},
			{"async", "Async function"},
			{"await", "Await expression"},
		}

		for _, k := range jsKeywords {
			if strings.HasPrefix(k.Text, prefix) {
				completions = append(completions, k)
			}
		}
	}

	return completions
}

// Show completions in a popup
func (e *Editor) showCompletions() {
	completions := e.getCompletions()
	if len(completions) == 0 {
		return
	}

	// Store completions for selection
	e.completions = completions
	e.completionIndex = 0
	e.completionActive = true

	// Draw completions (will be handled in Draw method)
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
	for wordStart > 0 && isWordChar(rune(line[wordStart-1])) {
		wordStart--
	}

	// Replace the current word with the completion
	newLine := line[:wordStart] + completion.Text + line[e.cursorX:]
	e.addUndo(e.lines[e.cursorY])
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
