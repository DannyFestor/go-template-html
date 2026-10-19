//go:build e2e

package e2e_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAVisitorRegistersThroughTheBrowserAndLandsOnTheDashboard(t *testing.T) {
	t.Parallel()

	page := openPage(t, "/register")

	for field, value := range map[string]string{
		"#name":                  "Ada Lovelace",
		"#email":                 "ada@example.com",
		"#password":              "correct horse",
		"#password_confirmation": "correct horse",
	} {
		require.NoError(t, page.Locator(field).Fill(value), field)
	}

	require.NoError(t, page.Locator(`form button[type="submit"]`).Click())

	require.NoError(t, expect.Page(page.Page).ToHaveURL(regexp.MustCompile(`/dashboard$`)))
	require.NoError(t, expect.Locator(page.Locator("main h1")).ToHaveText("Dashboard"))
	require.NoError(t, expect.Locator(page.Locator("#user-menu-name")).ToHaveText("Ada Lovelace"))
	assert.Empty(t, page.errors.list())
}
