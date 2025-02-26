# Kiki's Text Editor

A modern, feature-rich text editor written in Go, inspired by Vim and modern editors.

## Features

- Vim-like modal editing (Normal, Insert, Command modes)
- Syntax highlighting for multiple languages
- File tree navigation
- Search and replace functionality
- Undo/Redo support
- Auto-completion
- Line numbers
- Word wrap
- File backups
- Memory optimization for large files

## Keyboard Shortcuts

### Global
- `Ctrl+C`: Quit
- `?`: Show help
- `Ctrl+S`: Save file

### Normal Mode
- `i`: Enter insert mode
- `:`: Enter command mode
- `/`: Search
- `n`: Next search result
- `N`: Previous search result
- `u`: Undo
- `Ctrl+R`: Redo

### Insert Mode
- `ESC`: Return to normal mode
- `Tab`: Auto-complete (when available)

### Command Mode
Commands:
- `:w`: Save file
- `:q`: Quit
- `:wq`: Save and quit
- `:number`: Toggle line numbers
- `:wrap`: Toggle word wrap
- `:syntax`: Toggle syntax highlighting
- `:find <text>`: Search for text
- `:replace <old> <new>`: Replace text
- `:line <number>`: Jump to line
- `:info`: Show file information

## Installation

### Prerequisites
- Go 1.16 or higher
- Git
- GCC or equivalent C compiler
- Terminal with Unicode support

### Method 1: Using Go Install
```bash
# Install directly using Go
go install github.com/tino-sv/kikis-text-editor@latest

# Verify installation
kiki-text-editor --version
```

### Method 2: Building from Source
```bash
# Clone the repository
git clone https://github.com/yourusername/kikis-text-editor.git

# Navigate to project directory
cd kikis-text-editor

# Install dependencies
go mod download

# Build the project
go build -o kiki-editor

# Optional: Install globally
sudo mv kiki-editor /usr/local/bin/
```

### Platform-Specific Instructions

#### Linux
```bash
# Install required dependencies
sudo apt-get update
sudo apt-get install gcc make git curl

# Set up Go environment
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### macOS
```bash
# Using Homebrew
brew install go git

# Set up Go environment
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc
source ~/.zshrc
```

#### Windows
1. Install Go from [golang.org](https://golang.org)
2. Install Git from [git-scm.com](https://git-scm.com)
3. Install MinGW or MSYS2 for GCC
4. Add Go and Git to your PATH
5. Use PowerShell or Git Bash for commands

## Troubleshooting

### Common Issues

1. **"command not found" after installation**
   ```bash
   # Add to PATH manually
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

2. **Terminal display issues**
   - Ensure your terminal supports Unicode
   - Try setting a compatible font (e.g., Hack, Fira Code)
   - Set TERM environment variable:
     ```bash
     export TERM=xterm-256color
     ```

3. **Syntax highlighting not working**
   - Verify the file extension is supported
   - Check if syntax highlighting is enabled:
     ```
     :syntax on
     ```
   - Ensure your terminal supports colors

4. **Performance issues with large files**
   - Try disabling syntax highlighting for large files
   - Increase available memory:
     ```bash
     export GOGC=50
     ```

5. **File tree navigation issues**
   - Ensure read permissions for directories
   - Try refreshing the file tree with `:tree-refresh`

### Error Messages

1. **"failed to initialize screen"**
   - Check terminal compatibility
   - Try running in a different terminal
   - Verify TERM environment variable

2. **"permission denied" when saving**
   - Check file permissions
   - Try running with sudo (if appropriate)
   - Verify write permissions in directory

3. **"invalid syntax highlighting definition"**
   - Delete syntax cache: `rm -rf ~/.cache/kiki-editor/`
   - Reinstall the editor
   - Update to latest version

### Debug Mode

Run the editor in debug mode to get more information:
```bash
kiki-editor --debug file.txt
```

Debug logs are written to: `~/.kiki-editor/debug.log`

### Getting Help

1. Check the built-in help: Press `?` in normal mode
2. Run: `kiki-editor --help`
3. Visit our [GitHub Issues](https://github.com/tino-sv/kikis-text-editor/issues)
4. Join our [Discord community](https://discord.gg/sgerFXVj5e)

## Configuration

The editor can be configured through `~/.kiki_editor.json`. Example configuration:

```json
{
    "tabSize": 4,
    "showLineNumbers": true,
    "syntaxHighlight": true,
    "autoIndent": true,
    "wordWrap": false
}
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

## Contributing

### Getting Started

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Development Setup

```bash
# Clone your fork
git clone https://github.com/tino-sv/kikis-text-editor.git

# Add upstream remote
git remote add upstream https://github.com/tino-sv/kikis-text-editor.git

# Install development dependencies
go mod download

# Run tests
go test ./...

# Build and run locally
go run .
```

### Code Style

- Follow Go best practices and style guidelines
- Write tests for new features
- Update documentation as needed
- Keep commits atomic and well-described
- Add comments for complex logic

### Pull Request Process

1. Ensure all tests pass
2. Update the README.md with details of changes if applicable
3. Update the version number following [SemVer](http://semver.org/)
4. The PR will be merged once you have the sign-off of two maintainers

## License

MIT License

Copyright (c) 2024 Tino Valentine

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.