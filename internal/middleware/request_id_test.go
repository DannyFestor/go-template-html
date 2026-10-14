package middleware_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/middleware"
)

func TestRequestIDTagsTheResponseAndEveryLogLineWithANewUUIDv4(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer

	logger := slog.New(middleware.NewRequestIDLogHandler(slog.NewJSONHandler(&buffer, nil))).With("area", "test")
	handler := middleware.RequestID(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		logger.InfoContext(r.Context(), "handled")
	}))

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	request.Header.Set("X-Request-ID", "chosen-by-the-client")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)

	requestID := recorder.Header().Get("X-Request-ID")
	parsed, err := uuid.Parse(requestID)
	require.NoError(t, err)
	assert.Equal(t, byte(4), uuidVersion(parsed))

	var line map[string]any

	require.NoError(t, json.Unmarshal(buffer.Bytes(), &line))
	assert.Equal(t, requestID, line["request_id"])
	assert.Equal(t, "test", line["area"])
}

func uuidVersion(id uuid.UUID) byte {
	const versionOctet = 6

	return id[versionOctet] >> 4
}

func TestRequestIDLogHandlerLeavesLinesOutsideARequestAlone(t *testing.T) {
	t.Parallel()

	var buffer bytes.Buffer

	slog.New(middleware.NewRequestIDLogHandler(slog.NewJSONHandler(&buffer, nil))).InfoContext(t.Context(), "started")

	var line map[string]any

	require.NoError(t, json.Unmarshal(buffer.Bytes(), &line))
	assert.NotContains(t, line, "request_id")
}
