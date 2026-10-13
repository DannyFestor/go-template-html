package comment

import "strings"

type quote struct {
	mark      string
	escapes   bool
	multiline bool
}

type block struct {
	open, close string
}

// A syntax is what a lexer needs to find a language's comments without parsing it, so fragments work too.
type syntax struct {
	lineMarkers []string
	blocks      []block
	quotes      []quote
	directives  []func(comment string) bool
}

func (s syntax) isDirective(comment string) bool {
	if strings.Contains(comment, "gitleaks:allow") {
		return true
	}

	for _, directive := range s.directives {
		if directive(comment) {
			return true
		}
	}

	return false
}
