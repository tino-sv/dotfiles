package editor

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// Search functionality
func (e *Editor) startSearch() {
	e.mode = "search"
	e.searchTerm = ""
	e.searchMatches = nil
	e.currentMatch = 0
	e.setStatusMessage("Search: ")
}

func (e *Editor) handleSearchMode(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEnter:
		e.findMatches()
		e.mode = "normal"
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(e.searchTerm) > 0 {
			e.searchTerm = e.searchTerm[:len(e.searchTerm)-1]
		}
	case tcell.KeyRune:
		e.searchTerm += string(ev.Rune())
	}
	e.setStatusMessage(fmt.Sprintf("Search: %s", e.searchTerm))
}

func (e *Editor) findMatches() {
	e.searchMatches = nil
	for y, line := range e.lines {
		for x := 0; x < len(line); x++ {
			if strings.HasPrefix(line[x:], e.searchTerm) {
				e.searchMatches = append(e.searchMatches, struct{ y, x int }{y, x})
			}
		}
	}
	if len(e.searchMatches) > 0 {
		e.currentMatch = 0
		match := e.searchMatches[0]
		e.cursorY = match.y
		e.cursorX = match.x
		e.setStatusMessage(fmt.Sprintf("Match %d of %d", e.currentMatch+1, len(e.searchMatches)))
	} else {
		e.setStatusMessage("No matches found")
	}
}

func (e *Editor) nextMatch() {
	if len(e.searchMatches) == 0 {
		return
	}
	e.currentMatch = (e.currentMatch + 1) % len(e.searchMatches)
	match := e.searchMatches[e.currentMatch]
	e.cursorY = match.y
	e.cursorX = match.x
	e.setStatusMessage(fmt.Sprintf("Match %d of %d", e.currentMatch+1, len(e.searchMatches)))
}

func (e *Editor) previousMatch() {
	if len(e.searchMatches) == 0 {
		return
	}
	e.currentMatch--
	if e.currentMatch < 0 {
		e.currentMatch = len(e.searchMatches) - 1
	}
	match := e.searchMatches[e.currentMatch]
	e.cursorY = match.y
	e.cursorX = match.x
	e.setStatusMessage(fmt.Sprintf("Match %d of %d", e.currentMatch+1, len(e.searchMatches)))
}

func (e *Editor) findNext() bool {
	// Search from current position to end
	for y := e.cursorY; y < len(e.lines); y++ {
		line := e.lines[y]
		var x int
		if y == e.cursorY {
			x = e.cursorX + 1
		}

		if x < len(line) {
			if pos := strings.Index(line[x:], e.searchTerm); pos >= 0 {
				e.cursorY = y
				e.cursorX = x + pos
				return true
			}
		}
	}

	// Search from start to current position
	for y := 0; y < e.cursorY; y++ {
		if pos := strings.Index(e.lines[y], e.searchTerm); pos >= 0 {
			e.cursorY = y
			e.cursorX = pos
			return true
		}
	}

	return false
}

type SearchIndex struct {
	positions map[string][]Position
	dirty     bool
}

type Position struct {
	line, col int
}

func (e *Editor) buildSearchIndex() {
	if !e.searchIndex.dirty {
		return
	}

	e.searchIndex.positions = make(map[string][]Position)
	for y, line := range e.lines {
		words := strings.Fields(line)
		for _, word := range words {
			if len(word) > 2 { // Only index words longer than 2 chars
				e.searchIndex.positions[word] = append(
					e.searchIndex.positions[word],
					Position{y, strings.Index(line, word)},
				)
			}
		}
	}
	e.searchIndex.dirty = false
}
