package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/middleware"
)

func TestRecoverRendersAPanicAsAnError(t *testing.T) {
	t.Parallel()

	renderer := &spyRenderer{rendered: nil}
	handler := middleware.Recover(renderer)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("nil map write")
	}))

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusTeapot, recorder.Code)
	require.Len(t, renderer.rendered, 1)
	assert.ErrorContains(t, renderer.rendered[0], "nil map write")
}

func TestRecoverLetsAnAbortedHandlerAbort(t *testing.T) {
	t.Parallel()

	handler := middleware.Recover(
		&spyRenderer{rendered: nil},
	)(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic(http.ErrAbortHandler)
		}),
	)

	assert.PanicsWithValue(t, http.ErrAbortHandler, func() {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))
	})
}
