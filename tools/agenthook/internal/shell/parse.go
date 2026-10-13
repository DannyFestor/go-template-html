package shell

import (
	"errors"
	"fmt"
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

var errUnresolved = errors.New("command substitution not resolved")

type Start struct {
	Dir     string
	Environ []string
	// Substitute answers a command substitution without running it, because nothing may run before the script is
	// judged.
	Substitute func(Command) (string, bool)
}

func Parse(script string, start Start) ([]Command, error) {
	file, err := syntax.NewParser().Parse(strings.NewReader(script), "")
	if err != nil {
		return nil, fmt.Errorf("parse shell command: %w", err)
	}

	walker := newWalker(start)
	syntax.Walk(file, walker.visit)

	return walker.commands, walker.err
}
