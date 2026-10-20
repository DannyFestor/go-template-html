package validate

import (
	"net/mail"
	"strings"
	"unicode/utf8"
)

func NotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func EmailAddress(value string) bool {
	parsed, err := mail.ParseAddress(value)
	isBareAddress := err == nil && parsed.Address == value

	return isBareAddress
}

func MinRunes(value string, minimum int) bool {
	return utf8.RuneCountInString(value) >= minimum
}
