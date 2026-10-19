package health_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	healthhandler "github.com/DannyFestor/go-template-html/internal/handler/health"
)

func TestLivenessReportsOK(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()

	healthhandler.Liveness(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil))

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
	assert.Equal(t, "ok\n", recorder.Body.String())
}
