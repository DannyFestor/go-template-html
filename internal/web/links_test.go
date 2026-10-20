package web_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/web"
)

func linksOn(t *testing.T, base string) *web.Links {
	t.Helper()

	parsed, err := url.Parse(base)
	require.NoError(t, err)

	return web.NewLinks(*parsed)
}

func TestThePasswordResetURLCarriesTheTokenInItsQuery(t *testing.T) {
	t.Parallel()

	assert.Equal(t,
		"https://example.com/reset-password?token=RESET+TOKEN%2F1",
		linksOn(t, "https://example.com").PasswordResetURL("RESET TOKEN/1"),
	)
}

func TestThePasswordResetURLBuildsOnTheAppURLsPath(t *testing.T) {
	t.Parallel()

	assert.Equal(t,
		"https://example.com/app/reset-password?token=RESETTOKEN",
		linksOn(t, "https://example.com/app/").PasswordResetURL("RESETTOKEN"),
	)
}
