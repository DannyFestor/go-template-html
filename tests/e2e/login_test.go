//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/factory"
	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/testsupport"
)

func (a *app) createAda(t *testing.T) domain.User {
	t.Helper()

	return testsupport.Must(a.factory().User().WithEmail("ada@example.com").Create(t.Context()))(t)
}

func (a *app) openSignedIn(t *testing.T, user domain.User, path string) browserPage {
	t.Helper()

	page := a.open(t, route.Login)
	page.logIn(t, user)

	_, err := page.Goto(a.server.URL + path)
	require.NoError(t, err)

	return page
}

func (p browserPage) logIn(t *testing.T, user domain.User) {
	t.Helper()

	p.logInWith(t, user.Email, factory.Password)
}

func (p browserPage) logInWith(t *testing.T, email, password string) {
	t.Helper()

	require.NoError(t, p.Locator("#email").Fill(email))
	require.NoError(t, p.Locator("#password").Fill(password))
	p.clickTo(t, `form button[type="submit"]`, route.Dashboard)
}

func TestAUserLogsInThroughTheBrowserAndLogsOutFromTheUserMenu(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	ada := running.createAda(t)
	page := running.open(t, route.Login)

	page.logIn(t, ada)

	require.NoError(t, expect.Locator(page.Locator("#user-menu-name")).ToHaveText(ada.Name))

	page.click(t, "#user-menu-name")
	page.clickTo(t, `form[action="/logout"] button`, route.Home)

	require.NoError(t, expect.Locator(page.Locator("main h1")).ToHaveText("Let's get started"))
	assert.Empty(t, page.errors.list())
}
