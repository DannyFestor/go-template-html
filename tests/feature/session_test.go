//go:build feature

package feature_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAFlashSurvivesTheRedirectAndShowsOnce(t *testing.T) {
	t.Parallel()

	running := startApp(t)

	redirected := running.send(t, outgoing{method: http.MethodPost, path: probeFlash, headers: nil, body: nil})
	require.Equal(t, http.StatusSeeOther, redirected.status)

	target := redirected.header.Get("Location")
	assert.Equal(t, probeFlashMessage, running.page(t, target).Find(`main#public-main [role="status"]`).Text())
	assert.Zero(t, running.page(t, target).Find(`[role="status"]`).Length(), "the flash showed twice")
}

func TestASessionIsInvisibleToAnAppUnderAnotherKeyPrefix(t *testing.T) {
	t.Parallel()

	flashed, other := startApp(t), startApp(t)
	flashed.send(t, outgoing{method: http.MethodPost, path: probeFlash, headers: nil, body: nil})

	cookies := flashed.cookies(t)
	require.Len(t, cookies, 1)

	fetched := other.send(t, outgoing{
		method:  http.MethodGet,
		path:    "/",
		headers: map[string]string{"Cookie": cookies[0].String()},
		body:    nil,
	})

	assert.Zero(t, fetched.document(t).Find(`[role="status"]`).Length())
	assert.Equal(t, probeFlashMessage, flashed.page(t, "/").Find(`[role="status"]`).Text())
}
