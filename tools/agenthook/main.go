package main

import (
	"fmt"
	"os"

	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/claude"
)

// Exit status 2 would make Claude Code block the tool call, so every failure exits with 1.
func main() {
	events := os.Args[1:]
	if len(events) != 1 {
		fmt.Fprintln(os.Stderr, "usage: agenthook pre-tool-use|post-tool-use|stop")
		os.Exit(1)
	}

	err := claude.Run(events[0], os.Stdin, os.Stdout, os.Environ())
	if err != nil {
		fmt.Fprintln(os.Stderr, "agenthook:", err)
		os.Exit(1)
	}
}
