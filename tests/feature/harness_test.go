//go:build feature

package feature_test

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/bootstrap"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testdb"
)

const appName = "Feature Test App"

type app struct {
	server *httptest.Server
	logs   *lockedBuffer
}

type response struct {
	status           int
	header           http.Header
	closesConnection bool
	body             []byte
}

type outgoing struct {
	method  string
	path    string
	headers map[string]string
	body    io.Reader
}

type lockedBuffer struct {
	mutex  sync.Mutex
	buffer bytes.Buffer
}

func (b *lockedBuffer) Write(content []byte) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	written, err := b.buffer.Write(content)
	if err != nil {
		return written, fmt.Errorf("buffer the log line: %w", err)
	}

	return written, nil
}

func (b *lockedBuffer) String() string {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.buffer.String()
}

func startApp(t *testing.T) *app {
	t.Helper()

	return startAppOn(t, databaseConfig(testdb.New(t)))
}

func databaseConfig(database testdb.Database) bootstrap.DB {
	return bootstrap.DB{
		Host:                     database.Host,
		Port:                     database.Port,
		User:                     database.User,
		Password:                 bootstrap.Secret(database.Password),
		Name:                     database.Name,
		Schema:                   database.Schema,
		IdleInTransactionTimeout: time.Minute,
		MaxConnLifetime:          time.Hour,
		MaxConnLifetimeJitter:    time.Minute,
		MaxConnIdleTime:          time.Minute,
		ConnectTimeout:           time.Second,
		Web:                      bootstrap.WebDB{MaxConns: 4, MinConns: 0, StatementTimeout: 5 * time.Second},
	}
}

func startAppOn(t *testing.T, database bootstrap.DB) *app {
	t.Helper()

	config := bootstrap.Config{
		App: bootstrap.App{Name: appName, Env: bootstrap.EnvironmentProduction, ShutdownTimeout: time.Second},
		HTTP: bootstrap.HTTP{
			Addr:              "",
			ReadHeaderTimeout: time.Second,
			ReadTimeout:       time.Second,
			WriteTimeout:      time.Second,
			IdleTimeout:       time.Second,
			MaxBodyBytes:      maxBodyBytes,
			ReadinessCacheTTL: time.Second,
		},
		DB:  database,
		Log: bootstrap.Log{Level: slog.LevelInfo},
	}

	logs := &lockedBuffer{mutex: sync.Mutex{}, buffer: bytes.Buffer{}}

	webApp, err := bootstrap.NewWebApp(t.Context(), config, bootstrap.NewLogger(logs, config.App.Env, config.Log))
	require.NoError(t, err)
	t.Cleanup(webApp.Close)

	mountProbes(webApp)

	server := httptest.NewServer(webApp.Handler())
	t.Cleanup(server.Close)

	server.Client().CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }

	return &app{server: server, logs: logs}
}

func (a *app) get(t *testing.T, path string) response {
	t.Helper()

	return a.send(t, outgoing{method: http.MethodGet, path: path, headers: nil, body: nil})
}

func (a *app) send(t *testing.T, out outgoing) response {
	t.Helper()

	request, err := http.NewRequestWithContext(t.Context(), out.method, a.server.URL+out.path, out.body)
	require.NoError(t, err)

	for name, value := range out.headers {
		request.Header.Set(name, value)
	}

	received, err := a.server.Client().Do(request)
	require.NoError(t, err)

	defer func() { _ = received.Body.Close() }()

	body, err := io.ReadAll(received.Body)
	require.NoError(t, err)

	return response{
		status:           received.StatusCode,
		header:           received.Header,
		closesConnection: received.Close,
		body:             body,
	}
}

func (a *app) page(t *testing.T, path string) *goquery.Document {
	t.Helper()

	fetched := a.get(t, path)
	require.Equal(t, http.StatusOK, fetched.status)

	return fetched.document(t)
}

func (r response) document(t *testing.T) *goquery.Document {
	t.Helper()

	require.Equal(t, "text/html; charset=utf-8", r.header.Get("Content-Type"))

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(r.body))
	require.NoError(t, err)

	return document
}
