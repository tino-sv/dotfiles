## this is my first big solo project


# Text Editor

A modern, feature-rich text editor written entirely in Go, inspired by Vim.

## Features

## installation

### Prerequisites
- GO 1.16 or higher
- Git

## building from source

```bash
git clone https://github.com/tino-sv/kikis-text-editor.git
cd kikis-text-editor
go build or go install (for global installation)
```
## note: 'go install' will install the editor to your go path, so you can use it from anywhere.

## running the editor

```bash
./text-editor (will only work within the directory you cloned the repo into)
```



## Features

- CLI menu for file operations
- Vim-style modal editing (Normal and Insert modes)
- Command mode with common operations
- Syntax highlighting for code
- Search functionality
- Undo/Redo support
- Auto-closing brackets and quotes
- Line numbers
- Status bar with file info

### Editor Commands

#### Normal Mode
- `i` - Enter insert mode
- `ESC` - Return to normal mode
- `h,j,k,l` - Move cursor (left, down, up, right)
- `/` - Start search
- `n` - Next search match
- `N` - Previous search match
- `u` - Undo
- `r` - Redo

#### Command Mode (accessed with `:`)
- `:w` - Save file
- `:q` - Quit (if no unsaved changes)
- `:q!` - Force quit
- `:wq` - Save and quit
- `:set number` - Show line numbers
- `:set nonumber` - Hide line numbers

#### Insert Mode
- Type normally to insert text
- Auto-closes brackets and quotes
- `TAB` inserts spaces (configurable width)
- `ESC` to return to normal mode

### Status Bar Information
- Current mode (NORMAL/INSERT)
- Filename
- [+] indicator for unsaved changes
- Line and column numbers
- Search results
- Command feedback

## Project Structure

```
text-editor/
├── main.go           # Entry point
├── cli.go           # CLI menu interface
└── editor/          # Editor package
    ├── editor.go    # Core editor structure
    ├── input.go     # Input handling
    ├── display.go   # Screen rendering
    ├── cursor.go    # Cursor operations
    ├── search.go    # Search functionality
    ├── commands.go  # Command handling
    └── history.go   # Undo/redo functionality
```

## Building

```bash
go build
```

## Dependencies

- github.com/gdamore/tcell/v2 - Terminal handling