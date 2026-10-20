package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/middleware"
)

func TestNoReferrerOverridesTheDefaultReferrerPolicy(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	handler := middleware.SecurityHeaders(false)(middleware.NoReferrer(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }),
	))

	handler.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	assert.Equal(t, []string{"no-referrer"}, recorder.Header().Values("Referrer-Policy"))
}
