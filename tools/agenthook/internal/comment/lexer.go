package comment

import (
	"strings"
	"unicode"
)

type lexer struct {
	syntax   syntax
	text     string
	pos      int
	comments []string
}

func (s syntax) comments(text string) []string {
	lexer := lexer{syntax: s, text: text, pos: 0, comments: nil}
	for lexer.pos < len(text) {
		lexer.step()
	}

	return lexer.comments
}

func (l *lexer) step() {
	rest := l.text[l.pos:]

	for _, quote := range l.syntax.quotes {
		if strings.HasPrefix(rest, quote.mark) {
			l.skipQuoted(quote)

			return
		}
	}

	for _, block := range l.syntax.blocks {
		if strings.HasPrefix(rest, block.open) {
			l.readUntil(block.close, len(block.open))

			return
		}
	}

	for _, marker := range l.syntax.lineMarkers {
		if strings.HasPrefix(rest, marker) && l.atWordStart() {
			l.readUntil("\n", len(marker))

			return
		}
	}

	l.pos++
}

// A line marker inside a word, as in a URL's "//" or the shell's "${#x}", does not start a comment.
func (l *lexer) atWordStart() bool {
	return l.pos == 0 || unicode.IsSpace(rune(l.text[l.pos-1]))
}

func (l *lexer) readUntil(end string, markerLength int) {
	start := l.pos
	length := strings.Index(l.text[start+markerLength:], end)

	l.pos = len(l.text)
	if length >= 0 {
		l.pos = start + markerLength + length + len(end)
	}

	l.comments = append(l.comments, strings.TrimSpace(l.text[start:l.pos]))
}

func (l *lexer) skipQuoted(quote quote) {
	l.pos += len(quote.mark)

	for l.pos < len(l.text) {
		switch {
		case quote.escapes && l.text[l.pos] == '\\':
			l.pos += 2
		case strings.HasPrefix(l.text[l.pos:], quote.mark):
			l.pos += len(quote.mark)

			return
		case !quote.multiline && l.text[l.pos] == '\n':
			return
		default:
			l.pos++
		}
	}
}
