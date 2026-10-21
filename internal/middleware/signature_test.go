package middleware_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/middleware"
	"github.com/DannyFestor/go-template-html/internal/web"
)

var errBadSignature = errors.New("bad signature")

type acceptingSignedValue struct{}

func (acceptingSignedValue) Verify(target *url.URL) error {
	if target.Query().Get("signature") != "valid" {
		return errBadSignature
	}

	return nil
}

func requestTheSignedPage(t *testing.T, target string) (*httptest.ResponseRecorder, *spyRenderer) {
	t.Helper()

	renderer := &spyRenderer{rendered: nil}
	handler := middleware.RequireValidSignature(acceptingSignedValue{}, renderer)(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "Signed") }),
	)

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil))

	return recorder, renderer
}

func TestARequestWithAValidSignaturePasses(t *testing.T) {
	t.Parallel()

	recorder, renderer := requestTheSignedPage(t, "/verify-email/1/abc?signature=valid")

	assert.Equal(t, "Signed", recorder.Body.String())
	assert.Empty(t, renderer.rendered)
}

func TestARequestWithAnInvalidSignatureIsForbidden(t *testing.T) {
	t.Parallel()

	recorder, renderer := requestTheSignedPage(t, "/verify-email/1/abc?signature=forged")

	assert.Empty(t, recorder.Body.String())
	require.Len(t, renderer.rendered, 1)
	assert.ErrorIs(t, renderer.rendered[0], web.ErrForbidden)
}
