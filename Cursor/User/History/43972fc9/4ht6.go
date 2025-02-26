package editor

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	TabSize         int  `json:"tabSize"`
	ShowLineNumbers bool `json:"showLineNumbers"`
	SyntaxHighlight bool `json:"syntaxHighlight"`
	WordWrap        bool `json:"wordWrap"`
}

func (e *Editor) loadConfig() error {
	configPath := filepath.Join(os.Getenv("HOME"), ".kiki_editor.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return e.saveDefaultConfig()
		}
		return err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return err
	}

	// Apply config
	e.tabSize = config.TabSize
	e.showLineNumbers = config.ShowLineNumbers
	e.syntaxHighlight = config.SyntaxHighlight
	e.wordWrap = config.WordWrap

	return nil
}

func (e *Editor) saveDefaultConfig() error {
	config := Config{
		TabSize:         4,
		ShowLineNumbers: true,
		SyntaxHighlight: true,
		WordWrap:        false,
	}

	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	configPath := filepath.Join(os.Getenv("HOME"), ".kiki_editor.json")
	return os.WriteFile(configPath, data, 0644)
}
