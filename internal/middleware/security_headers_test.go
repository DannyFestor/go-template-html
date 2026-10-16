package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/middleware"
)

func securityHeadersOf(t *testing.T, strictTransport bool) http.Header {
	t.Helper()

	recorder := httptest.NewRecorder()
	handler := middleware.SecurityHeaders(
		strictTransport,
	)(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}),
	)

	handler.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	return recorder.Header()
}

func TestSecurityHeadersLockDownEveryResponse(t *testing.T) {
	t.Parallel()

	header := securityHeadersOf(t, false)

	assert.Equal(t,
		"default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; font-src 'self'; "+
			"connect-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'; object-src 'none'",
		header.Get("Content-Security-Policy"))
	assert.Equal(t, "nosniff", header.Get("X-Content-Type-Options"))
	assert.Equal(t, "strict-origin-when-cross-origin", header.Get("Referrer-Policy"))
	assert.Equal(t, "same-origin", header.Get("Cross-Origin-Opener-Policy"))
	assert.Equal(t, "camera=(), microphone=(), geolocation=(), payment=(), usb=()", header.Get("Permissions-Policy"))
	assert.Empty(t, header.Get("Strict-Transport-Security"))
}

func TestSecurityHeadersPinHTTPSWhenTheAppIsServedOverIt(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "max-age=31536000", securityHeadersOf(t, true).Get("Strict-Transport-Security"))
}
