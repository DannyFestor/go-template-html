package shell

import (
	"slices"
	"strconv"

	"mvdan.cc/sh/v3/syntax"
)

type redirects struct {
	writes []string
	input  string
}

func (e *expander) redirects(redirs []*syntax.Redirect, dir string) redirects {
	var result redirects

	for _, redir := range redirs {
		switch {
		case writesFile(redir):
			result.writes = appendResolved(result.writes, dir, e.word(redir.Word))
		case redir.Op == syntax.Hdoc || redir.Op == syntax.DashHdoc:
			result.input = e.heredoc(redir.Hdoc)
		case redir.Op == syntax.WordHdoc:
			result.input = e.word(redir.Word)
		}
	}

	return result
}

func writesFile(redir *syntax.Redirect) bool {
	if redir.Op == syntax.DplOut {
		return !duplicatesDescriptor(redir.Word.Lit())
	}

	return slices.Contains([]syntax.RedirOperator{
		syntax.RdrOut, syntax.AppOut, syntax.RdrInOut, syntax.RdrClob, syntax.AppClob,
		syntax.RdrAll, syntax.RdrAllClob, syntax.AppAll, syntax.AppAllClob,
	}, redir.Op)
}

func duplicatesDescriptor(target string) bool {
	_, err := strconv.Atoi(target)

	return target == "-" || err == nil
}

func appendResolved(paths []string, dir, path string) []string {
	resolved := Resolve(dir, path)
	if resolved == "" {
		return paths
	}

	return append(paths, resolved)
}
