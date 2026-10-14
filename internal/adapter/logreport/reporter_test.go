package logreport_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/logreport"
)

var errDiskFull = errors.New("disk full")

func TestReportLogsTheErrorWithItsStack(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer

	reporter := logreport.New(slog.New(slog.NewJSONHandler(&buffer, nil)))
	reporter.Report(t.Context(), errDiskFull)

	var line map[string]any

	require.NoError(t, json.Unmarshal(buffer.Bytes(), &line))
	assert.Equal(t, "ERROR", line["level"])
	assert.Equal(t, "disk full", line["error"])
	assert.Contains(t, line["stack"], "TestReportLogsTheErrorWithItsStack")
}
