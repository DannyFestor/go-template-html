package bootstrap_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/bootstrap"
)

func TestNewLoggerWritesJSONInProduction(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer

	bootstrap.NewLogger(&buffer, bootstrap.EnvironmentProduction, bootstrap.Log{Level: slog.LevelInfo}).Info("started")

	var line map[string]any

	require.NoError(t, json.Unmarshal(buffer.Bytes(), &line))
	assert.Equal(t, "started", line["msg"])
}

func TestNewLoggerWritesTextLocally(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer

	bootstrap.NewLogger(&buffer, bootstrap.EnvironmentLocal, bootstrap.Log{Level: slog.LevelInfo}).Info("started")

	assert.Contains(t, buffer.String(), "level=INFO msg=started")
}

func TestNewLoggerDropsRecordsBelowTheLevel(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer

	logger := bootstrap.NewLogger(&buffer, bootstrap.EnvironmentLocal, bootstrap.Log{Level: slog.LevelWarn})
	logger.Info("dropped")
	logger.Warn("kept")

	assert.NotContains(t, buffer.String(), "dropped")
	assert.Contains(t, buffer.String(), "kept")
}
