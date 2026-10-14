package bootstrap_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/bootstrap"
)

const plaintext = "hunter2"

type holder struct {
	Field bootstrap.Secret `json:"field"`
}

func TestSecretIsRedactedInFormattedOutput(t *testing.T) {
	t.Parallel()

	secret := bootstrap.Secret(plaintext)

	for _, verb := range []string{"%v", "%+v", "%#v", "%s", "%q", "%x", "%X", "%d", "%T %[1]v"} {
		t.Run(verb, func(t *testing.T) {
			t.Parallel()

			assert.NotContains(t, fmt.Sprintf(verb, secret), plaintext)
			assert.NotContains(t, fmt.Sprintf(verb, holder{Field: secret}), plaintext)
		})
	}
}

func TestSecretIsRedactedInLogs(t *testing.T) {
	t.Parallel()

	handlers := map[string]func(*bytes.Buffer) slog.Handler{
		"json": func(buffer *bytes.Buffer) slog.Handler { return slog.NewJSONHandler(buffer, nil) },
		"text": func(buffer *bytes.Buffer) slog.Handler { return slog.NewTextHandler(buffer, nil) },
	}

	for name, newHandler := range handlers {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var buffer bytes.Buffer

			logger := slog.New(newHandler(&buffer))
			logger.Info("config", "secret", bootstrap.Secret(plaintext), "holder", holder{Field: plaintext})

			assert.NotContains(t, buffer.String(), plaintext)
			assert.Contains(t, buffer.String(), "[redacted]")
		})
	}
}

func TestSecretIsRedactedInJSON(t *testing.T) {
	t.Parallel()

	encoded, err := json.Marshal(holder{Field: plaintext})

	require.NoError(t, err)
	assert.JSONEq(t, `{"field":"[redacted]"}`, string(encoded))
}

func TestSecretRevealsItsValue(t *testing.T) {
	t.Parallel()

	assert.Equal(t, plaintext, bootstrap.Secret(plaintext).Reveal())
}
