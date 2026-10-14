package web_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/web"
)

func TestHandleLeavesASuccessfulResponseAlone(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	handler := newResponder(
		&spyReporter{reported: nil},
		false,
	).Handle(func(w http.ResponseWriter, _ *http.Request) error {
		_, _ = io.WriteString(w, "done")

		return nil
	})

	handler.ServeHTTP(recorder, newRequest(t, nil))

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "done", recorder.Body.String())
}

func TestHandleRendersTheReturnedErrorsPage(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	handler := newResponder(&spyReporter{reported: nil}, false).Handle(func(http.ResponseWriter, *http.Request) error {
		return web.ErrNotFound
	})

	handler.ServeHTTP(recorder, newRequest(t, nil))

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Not Found")
}
