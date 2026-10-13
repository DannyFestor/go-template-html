package comment

import (
	"regexp"
	"strings"
)

var (
	generatedHeader     = regexp.MustCompile(`(?m)^// Code generated .* DO NOT EDIT\.$`)
	goNameDirective     = regexp.MustCompile(`^//[a-z][a-z0-9_-]*:\S`)
	nolintWithReason    = regexp.MustCompile(`^//nolint:[\w,-]+ // \S`)
	typeScriptDirective = regexp.MustCompile(`^// @ts-expect-error\b|^/// <reference |` +
		`^(//|/\*)\s*oxlint-(disable|disable-line|disable-next-line|enable)\s+[a-z@]`)
	sqlDirective   = regexp.MustCompile(`^-- name: \S+ :\w+|^-- \+goose `)
	shellDirective = regexp.MustCompile(`^#!|^# shellcheck `)
)

func syntaxFor(extension string) (syntax, bool) {
	switch extension {
	case ".go":
		return goSyntax(), true
	case ".ts":
		return typeScriptSyntax(), true
	case ".templ":
		return templSyntax(), true
	case ".sql":
		return sqlSyntax(), true
	case ".sh":
		return shellSyntax(), true
	default:
		return syntax{lineMarkers: nil, blocks: nil, quotes: nil, directives: nil}, false
	}
}

func goSyntax() syntax {
	return syntax{
		lineMarkers: []string{"//"},
		blocks:      []block{cBlock()},
		quotes:      []quote{escaped(`"`), escaped(`'`), raw("`")},
		directives:  []func(string) bool{isGoDirective},
	}
}

func typeScriptSyntax() syntax {
	return syntax{
		lineMarkers: []string{"//"},
		blocks:      []block{cBlock()},
		quotes:      []quote{escaped(`"`), escaped(`'`), {mark: "`", escapes: true, multiline: true}},
		directives:  []func(string) bool{typeScriptDirective.MatchString},
	}
}

// Apostrophes are common in templ's HTML text, so only Go's double and back quotes delimit strings.
func templSyntax() syntax {
	return syntax{
		lineMarkers: []string{"//"},
		blocks:      []block{cBlock(), {open: "<!--", close: "-->"}},
		quotes:      []quote{escaped(`"`), raw("`")},
		directives:  []func(string) bool{isGoDirective},
	}
}

func sqlSyntax() syntax {
	return syntax{
		lineMarkers: []string{"--"},
		blocks:      []block{cBlock()},
		quotes:      []quote{raw(`'`), raw(`"`)},
		directives:  []func(string) bool{sqlDirective.MatchString},
	}
}

func shellSyntax() syntax {
	return syntax{
		lineMarkers: []string{"#"},
		blocks:      nil,
		quotes:      []quote{raw(`'`), {mark: `"`, escapes: true, multiline: true}},
		directives:  []func(string) bool{shellDirective.MatchString},
	}
}

func cBlock() block {
	return block{open: "/*", close: "*/"}
}

// An escaped quote ends at the end of its line, so a stray one in a fragment cannot hide the comments after it.
func escaped(mark string) quote {
	return quote{mark: mark, escapes: true, multiline: false}
}

func raw(mark string) quote {
	return quote{mark: mark, escapes: false, multiline: true}
}

// nolintlint rejects a //nolint without a reason, so only one with a reason counts as a directive.
func isGoDirective(comment string) bool {
	if strings.HasPrefix(comment, "//nolint") {
		return nolintWithReason.MatchString(comment)
	}

	return goNameDirective.MatchString(comment) || generatedHeader.MatchString(comment)
}

func HasGeneratedHeader(text string) bool {
	return generatedHeader.MatchString(text)
}
