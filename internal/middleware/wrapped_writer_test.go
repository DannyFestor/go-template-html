package middleware_test

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/middleware"
)

type spyRenderer struct {
	rendered []error
}

func (s *spyRenderer) Error(w http.ResponseWriter, _ *http.Request, err error) {
	s.rendered = append(s.rendered, err)

	w.WriteHeader(http.StatusTeapot)
}

type writerCapabilities struct {
	unwraps    bool
	flushes    bool
	hijackErr  error
	isHijacker bool
}

func TestWrappingMiddlewareKeepsTheWritersCapabilities(t *testing.T) {
	t.Parallel()

	wrappers := map[string]func(http.Handler) http.Handler{
		"access log":            middleware.AccessLog(slog.New(slog.DiscardHandler)),
		"mux error interceptor": middleware.InterceptMuxErrors(&spyRenderer{rendered: nil}),
	}

	for name, wrap := range wrappers {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			capabilities := make(chan writerCapabilities, 1)
			server := httptest.NewServer(wrap(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				capabilities <- inspect(w)
			})))
			t.Cleanup(server.Close)

			response, err := server.Client().Do(newServerRequest(t, server.URL))
			if err == nil {
				_ = response.Body.Close()
			}

			found := <-capabilities
			assert.True(t, found.unwraps, "Unwrap")
			assert.True(t, found.flushes, "Flush")
			assert.True(t, found.isHijacker, "Hijack")
			require.NoError(t, found.hijackErr)
		})
	}
}

func inspect(w http.ResponseWriter) writerCapabilities {
	_, unwraps := w.(interface{ Unwrap() http.ResponseWriter })
	_, flushes := w.(http.Flusher)
	hijacker, isHijacker := w.(http.Hijacker)

	found := writerCapabilities{unwraps: unwraps, flushes: flushes, hijackErr: nil, isHijacker: isHijacker}
	if isHijacker {
		conn, _, err := hijacker.Hijack()
		found.hijackErr = err

		if err == nil {
			_ = conn.Close()
		}
	}

	return found
}

func newServerRequest(t *testing.T, url string) *http.Request {
	t.Helper()

	request, err := http.NewRequestWithContext(context.WithoutCancel(t.Context()), http.MethodGet, url, nil)
	require.NoError(t, err)

	return request
}
