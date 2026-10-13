package shell

import (
	"strings"

	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/syntax"
)

type expander struct {
	config    *expand.Config
	variables []string
}

// Globbing and process substitution stay off, because nothing may run before the script is judged.
func newExpander(environ []string) *expander {
	config := new(expand.Config)
	config.Env = expand.ListEnviron(environ...)

	return &expander{config: config, variables: environ}
}

// An unknown word, such as a command substitution, expands to the empty string.
func (e *expander) word(word *syntax.Word) string {
	if word == nil {
		return ""
	}

	literal, err := expand.Literal(e.config, word)
	if err != nil {
		return ""
	}

	return literal
}

func (e *expander) words(words []*syntax.Word) []string {
	literals := make([]string, 0, len(words))
	for _, word := range words {
		literals = append(literals, e.word(word))
	}

	return literals
}

// The heredoc body is kept as written, because quoted delimiters turn expansion off.
func (e *expander) heredoc(word *syntax.Word) string {
	var body strings.Builder

	err := syntax.NewPrinter().Print(&body, word)
	if err != nil {
		return ""
	}

	return body.String()
}

func (e *expander) assign(assigns []*syntax.Assign) {
	for _, assign := range assigns {
		if !assign.Naked {
			e.variables = append(e.variables, assign.Name.Value+"="+e.word(assign.Value))
		}
	}

	e.config.Env = expand.ListEnviron(e.variables...)
}

func (e *expander) keepFirstVariables(count int) {
	e.variables = e.variables[:count]
	e.config.Env = expand.ListEnviron(e.variables...)
}

func (e *expander) cdTarget(command Command) string {
	for _, arg := range command.Args[1:] {
		if arg == "-" || !strings.HasPrefix(arg, "-") {
			return Resolve(command.Dir, arg)
		}
	}

	return e.config.Env.Get("HOME").String()
}
