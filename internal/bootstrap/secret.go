package bootstrap

import (
	"fmt"
	"io"
	"log/slog"
)

const redacted = "[redacted]"

type Secret string

func (s Secret) Reveal() string {
	return string(s)
}

func (Secret) String() string {
	return redacted
}

// Format covers the verbs that fmt does not route through String, such as %d and %#v, which would print the plaintext.
func (Secret) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, redacted)
}

func (Secret) LogValue() slog.Value {
	return slog.StringValue(redacted)
}

// MarshalText is needed because slog resolves LogValue only on the logged value itself, so a Secret nested in a
// struct reaches encoding/json.
func (Secret) MarshalText() ([]byte, error) {
	return []byte(redacted), nil
}
