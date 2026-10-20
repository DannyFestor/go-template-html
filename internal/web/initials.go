package web

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const initialsWords = 2

func initials(name string) string {
	var built strings.Builder

	for _, word := range firstWords(name, initialsWords) {
		first, _ := utf8.DecodeRuneInString(word)
		built.WriteRune(unicode.ToUpper(first))
	}

	return built.String()
}

func firstWords(text string, count int) []string {
	words := strings.Fields(text)

	return words[:min(count, len(words))]
}
