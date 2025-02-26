package main

import (
	"flag"
	"fmt"
	"os"
	"text-editor/editor"
)

const Version = "0.2.0.4"

func main() {
	// Parse command line flags
	debug := flag.Bool("debug", false, "Enable debug mode")
	version := flag.Bool("version", false, "Show version")
	help := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *version {
		fmt.Printf("Kiki's Text Editor v%s\n", Version)
		return
	}

	if *help {
		printHelp()
		return
	}

	ed, err := editor.NewEditor()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating editor: %v\n", err)
		os.Exit(1)
	}

	if *debug {
		ed.EnableDebugMode()
	}

	// If a file path is provided, open it
	args := flag.Args()
	if len(args) > 0 {
		ed.SetFilename(args[0])
		if err := ed.LoadFile(args[0]); err != nil {
			ed.SetStatusMessage(fmt.Sprintf("Error loading file: %v", err))
		}
	} else {
		ed.SetFilename("[New File]")
	}

	// Run the editor (will show file tree by default)
	ed.Run()
}

func printHelp() {
	fmt.Println(`Kiki's Text Editor

Usage: kiki-editor [options] [file]

Options:
  --debug     Enable debug mode
  --version   Show version
  --help      Show this help message

Commands in editor:
  Ctrl+C      Quit
  ?           Show help
  Ctrl+S      Save file
  :w          Save file
  :q          Quit
  :wq         Save and quit
  :number     Toggle line numbers
  :wrap       Toggle word wrap
  :syntax     Toggle syntax highlighting`)
}
