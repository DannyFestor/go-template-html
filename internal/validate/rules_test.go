package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/validate"
)

func TestNotBlank(t *testing.T) {
	t.Parallel()

	assert.True(t, validate.NotBlank("Ada"))
	assert.False(t, validate.NotBlank(""))
	assert.False(t, validate.NotBlank(" \t\n"))
}

func TestEmailAddress(t *testing.T) {
	t.Parallel()

	for email, valid := range map[string]bool{
		"ada@example.com":             true,
		"Ada.Lovelace+tag@Example.ac": true,
		"":                            false,
		"ada":                         false,
		"ada@":                        false,
		"@example.com":                false,
		"Ada <ada@example.com>":       false,
		"ada@example.com, b@b.com":    false,
	} {
		assert.Equal(t, valid, validate.EmailAddress(email), email)
	}
}

func TestMinRunesCountsCharactersNotBytes(t *testing.T) {
	t.Parallel()

	assert.True(t, validate.MinRunes("12345678", 8))
	assert.False(t, validate.MinRunes("1234567", 8))
	assert.True(t, validate.MinRunes("ääääääää", 8))
	assert.False(t, validate.MinRunes("äääää", 8), "ten bytes, five characters")
}
