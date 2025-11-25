package main

import (
	"github.com/asad/bluestack/internal/cli"
)

// main is the entry point for the Bluestack application.
// It delegates to the CLI package which handles command parsing and execution.
func main() {
	cli.Execute()
}

