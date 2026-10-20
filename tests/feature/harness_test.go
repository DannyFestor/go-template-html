//go:build feature

package feature_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/bootstrap"
	"github.com/DannyFestor/go-template-html/internal/factory"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testapp"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testdb"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testvalkey"
)

const (
	appName     = "Feature Test App"
	factorySeed = 1
)

type app struct {
	*testapp.App

	server    *httptest.Server
	logs      *lockedBuffer
	probeJobs *probeJobHandler
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

	return startAppOn(t, liveConfig(t))
}

func liveConfig(t *testing.T) bootstrap.Config {
	t.Helper()

	return testapp.Config(appName, testdb.New(t), testvalkey.New(t))
}

func startAppOn(t *testing.T, config bootstrap.Config) *app {
	t.Helper()

	logs := &lockedBuffer{mutex: sync.Mutex{}, buffer: bytes.Buffer{}}

	testApp := testapp.New(t, config, logs)
	mountProbes(testApp.WebApp)
	probeJobs := mountJobProbe(testApp)
	mountMailProbe(testApp)

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	server := testapp.Serve(t, testApp.Handler())
	server.Client().CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	server.Client().Jar = jar

	return &app{App: testApp, server: server, logs: logs, probeJobs: probeJobs}
}

func (a *app) cookies(t *testing.T) []*http.Cookie {
	t.Helper()

	return a.server.Client().Jar.Cookies(a.address(t))
}

func (a *app) address(t *testing.T) *url.URL {
	t.Helper()

	address, err := url.Parse(a.server.URL)
	require.NoError(t, err)

	return address
}

func (a *app) get(t *testing.T, path string) response {
	t.Helper()

	return a.send(t, outgoing{method: http.MethodGet, path: path, headers: nil, body: nil})
}

func (a *app) post(t *testing.T, path string, form url.Values) response {
	t.Helper()

	return a.send(t, outgoing{
		method:  http.MethodPost,
		path:    path,
		headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
		body:    strings.NewReader(form.Encode()),
	})
}

func (a *app) send(t *testing.T, out outgoing) response {
	t.Helper()

	return a.sendThrough(t, a.server.Client(), out)
}

func (a *app) sendWithoutCookieJar(t *testing.T, out outgoing) response {
	t.Helper()

	jarless := *a.server.Client()
	jarless.Jar = nil

	return a.sendThrough(t, &jarless, out)
}

func (a *app) sendThrough(t *testing.T, client *http.Client, out outgoing) response {
	t.Helper()

	request, err := http.NewRequestWithContext(t.Context(), out.method, a.server.URL+out.path, out.body)
	require.NoError(t, err)

	for name, value := range out.headers {
		request.Header.Set(name, value)
	}

	received, err := client.Do(request)
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

func (a *app) factory() *factory.Factory {
	return factory.New(a.Users, a.Passwords, factorySeed)
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
