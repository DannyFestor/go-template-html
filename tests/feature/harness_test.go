//go:build feature

package feature_test

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/bootstrap"
)

const appName = "Feature Test App"

type app struct {
	server *httptest.Server
}

type response struct {
	status int
	header http.Header
	body   []byte
}

func startApp(t *testing.T) *app {
	t.Helper()

	config := bootstrap.Config{
		App:  bootstrap.App{Name: appName, Env: bootstrap.EnvironmentProduction, ShutdownTimeout: time.Second},
		HTTP: bootstrap.HTTP{Addr: "", ReadHeaderTimeout: time.Second},
		Log:  bootstrap.Log{Level: slog.LevelInfo},
	}

	handler, err := bootstrap.NewWebHandler(config, slog.New(slog.DiscardHandler))
	require.NoError(t, err)

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return &app{server: server}
}

func (a *app) get(t *testing.T, path string) response {
	t.Helper()

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, a.server.URL+path, nil)
	require.NoError(t, err)

	received, err := a.server.Client().Do(request)
	require.NoError(t, err)

	defer func() { _ = received.Body.Close() }()

	body, err := io.ReadAll(received.Body)
	require.NoError(t, err)

	return response{status: received.StatusCode, header: received.Header, body: body}
}

func (a *app) page(t *testing.T, path string) *goquery.Document {
	t.Helper()

	fetched := a.get(t, path)
	require.Equal(t, http.StatusOK, fetched.status)
	require.Equal(t, "text/html; charset=utf-8", fetched.header.Get("Content-Type"))

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(fetched.body))
	require.NoError(t, err)

	return document
}
