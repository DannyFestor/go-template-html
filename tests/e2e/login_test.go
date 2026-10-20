//go:build e2e

package e2e_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/factory"
	"github.com/DannyFestor/go-template-html/internal/testsupport"
)

func TestAUserLogsInThroughTheBrowserAndLogsOutFromTheUserMenu(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	ada := testsupport.Must(running.factory().User().WithEmail("ada@example.com").Create(t.Context()))(t)
	page := running.open(t, "/login")

	require.NoError(t, page.Locator("#email").Fill(ada.Email))
	require.NoError(t, page.Locator("#password").Fill(factory.Password))
	require.NoError(t, page.Locator(`form button[type="submit"]`).Click())

	require.NoError(t, expect.Page(page.Page).ToHaveURL(regexp.MustCompile(`/dashboard$`)))
	require.NoError(t, expect.Locator(page.Locator("#user-menu-name")).ToHaveText(ada.Name))

	require.NoError(t, page.Locator("#user-menu-name").Click())
	require.NoError(t, page.Locator(`form[action="/logout"] button`).Click())

	require.NoError(t, expect.Page(page.Page).ToHaveURL(regexp.MustCompile(`/$`)))
	require.NoError(t, expect.Locator(page.Locator("main h1")).ToHaveText("Let's get started"))
	assert.Empty(t, page.errors.list())
}
