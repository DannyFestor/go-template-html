package compression_test

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/compression"
	"github.com/DannyFestor/go-template-html/internal/web"
)

const (
	html        = "text/html; charset=utf-8"
	minimumSize = 1024
)

func TestZstdWinsWhenTheClientAcceptsBothEncodingsEqually(t *testing.T) {
	t.Parallel()

	response := serve(t, writing(html, minimumSize), "gzip, zstd")

	assert.Equal(t, "zstd", response.Header().Get("Content-Encoding"))
	assert.Contains(t, response.Header().Values("Vary"), "Accept-Encoding")
}

func TestGzipServesAClientWithoutZstd(t *testing.T) {
	t.Parallel()

	response := serve(t, writing(html, minimumSize), "gzip")

	assert.Equal(t, "gzip", response.Header().Get("Content-Encoding"))
}

func TestEveryAllowListedContentTypeIsCompressed(t *testing.T) {
	t.Parallel()

	for _, contentType := range []string{
		html,
		"text/plain",
		"text/css",
		"text/javascript",
		"application/javascript",
		"application/json",
		"image/svg+xml",
		"application/xml",
	} {
		t.Run(contentType, func(t *testing.T) {
			t.Parallel()

			response := serve(t, writing(contentType, minimumSize), "gzip")

			assert.Equal(t, "gzip", response.Header().Get("Content-Encoding"))
		})
	}
}

func TestResponsesAreSentPlain(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		handler http.Handler
		size    int
	}{
		"below 1 KiB":     {handler: writing(html, minimumSize-1), size: minimumSize - 1},
		"an event stream": {handler: writing("text/event-stream", minimumSize), size: minimumSize},
		"an image":        {handler: writing("image/png", minimumSize), size: minimumSize},
		"already encoded": {
			handler: withHeader("Content-Encoding", "br", writing(html, minimumSize)),
			size:    minimumSize,
		},
		"a partial range": {
			handler: withHeader("Content-Range", "bytes 0-1023/2048", writing(html, minimumSize)),
			size:    minimumSize,
		},
	}

	for name, plain := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := serve(t, plain.handler, "gzip, zstd")

			assert.NotContains(t, []string{"gzip", "zstd"}, response.Header().Get("Content-Encoding"))
			assert.Equal(t, strings.Repeat("a", plain.size), response.Body.String())
		})
	}
}

func TestAResponseSkippingCompressionIsSentPlainWithoutTheMarker(t *testing.T) {
	t.Parallel()

	for _, acceptEncoding := range []string{"gzip, zstd", ""} {
		t.Run(acceptEncoding, func(t *testing.T) {
			t.Parallel()

			response := serve(t, skippingCompression(writing(html, minimumSize)), acceptEncoding)

			assert.Empty(t, response.Header().Get("Content-Encoding"))
			assert.Empty(t, response.Header().Values("No-Gzip-Compression"))
			assert.Equal(t, strings.Repeat("a", minimumSize), response.Body.String())
		})
	}
}

func TestAFlushSendsWhatWasWrittenSoFar(t *testing.T) {
	t.Parallel()

	flushed := make(chan string, 1)
	server := httptest.NewServer(wrap(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: first\n\n")
		assert.NoError(t, http.NewResponseController(w).Flush())
		<-flushed
	})))
	t.Cleanup(server.Close)

	response := get(t, server.URL)
	defer func() { _ = response.Body.Close() }()

	line, err := bufio.NewReader(response.Body).ReadString('\n')
	flushed <- line

	require.NoError(t, err)
	assert.Equal(t, "data: first\n", line)
}

func TestTheConnectionCanBeHijacked(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(wrap(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		conn, buffered, err := http.NewResponseController(w).Hijack()
		if !assert.NoError(t, err) {
			return
		}

		defer func() { _ = conn.Close() }()

		_, _ = buffered.WriteString("HTTP/1.1 101 Switching Protocols\r\n\r\n")
		_ = buffered.Flush()
	})))
	t.Cleanup(server.Close)

	var dialer net.Dialer

	conn, err := dialer.DialContext(t.Context(), "tcp", server.Listener.Addr().String())
	require.NoError(t, err)

	defer func() { _ = conn.Close() }()

	_, err = io.WriteString(conn, "GET / HTTP/1.1\r\nHost: test\r\nAccept-Encoding: gzip\r\n\r\n")
	require.NoError(t, err)

	status, err := bufio.NewReader(conn).ReadString('\n')
	require.NoError(t, err)
	assert.Equal(t, "HTTP/1.1 101 Switching Protocols\r\n", status)
}

func serve(t *testing.T, handler http.Handler, acceptEncoding string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	request.Header.Set("Accept-Encoding", acceptEncoding)

	recorder := httptest.NewRecorder()
	wrap(t, handler).ServeHTTP(recorder, request)

	return recorder
}

func wrap(t *testing.T, handler http.Handler) http.Handler {
	t.Helper()

	compress, err := compression.New()
	require.NoError(t, err)

	return compress(handler)
}

func get(t *testing.T, url string) *http.Response {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	t.Cleanup(cancel)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	require.NoError(t, err)
	request.Header.Set("Accept-Encoding", "gzip, zstd")

	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)

	return response
}

func writing(contentType string, size int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", contentType)
		_, _ = io.WriteString(w, strings.Repeat("a", size))
	})
}

func withHeader(name, value string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(name, value)
		next.ServeHTTP(w, r)
	})
}

func skippingCompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		web.SkipCompression(w)
		next.ServeHTTP(w, r)
	})
}
