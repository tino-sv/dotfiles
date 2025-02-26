package editor

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Default settings
//
//nolint:unused
var defaultSettings = map[string]string{
	"tabSize":         "4",
	"showLineNumbers": "true",
	"syntaxHighlight": "true",
	"autoIndent":      "true",
	"autoComplete":    "true",
	"theme":           "default",
	"saveOnExit":      "false",
	"backupFiles":     "true",
	"smartIndent":     "true",
	"wordWrap":        "false",
}

// Load settings from config file
func (e *Editor) loadSettings() {
	// Start with defaults
	for k, v := range defaultSettings {
		e.settings[k] = v
	}

	// Try to open config file
	file, err := os.Open(e.configFile)
	if err != nil {
		// Config file doesn't exist, create with defaults
		e.saveSettings()
		return
	}
	defer file.Close()

	// Read settings
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			e.settings[key] = value
		}
	}

	// Apply settings
	e.applySettings()
}

// Save settings to config file
func (e *Editor) saveSettings() {
	file, err := os.Create(e.configFile)
	if err != nil {
		e.setStatusMessage(fmt.Sprintf("Error saving settings: %v", err))
		return
	}
	defer file.Close()

	// Write header
	file.WriteString("# Kiki's Text Editor Configuration\n\n")

	// Write settings
	for k, v := range e.settings {
		file.WriteString(fmt.Sprintf("%s = %s\n", k, v))
	}

	e.setStatusMessage("Settings saved")
}

// Apply loaded settings to editor
func (e *Editor) applySettings() {
	// Apply tab size
	if val, ok := e.settings["tabSize"]; ok {
		var tabSize int
		fmt.Sscan(val, &tabSize)
		if tabSize > 0 {
			e.tabSize = tabSize
		}
	}

	// Apply line numbers
	if val, ok := e.settings["showLineNumbers"]; ok {
		e.showLineNumbers = (val == "true")
	}

	// Apply syntax highlighting
	if val, ok := e.settings["syntaxHighlight"]; ok {
		e.syntaxHighlight = (val == "true")
	}

	// Other settings can be applied as needed
}

// Set a setting and save
func (e *Editor) setSetting(key, value string) {
	e.settings[key] = value
	e.applySettings()
	e.saveSettings()
}
