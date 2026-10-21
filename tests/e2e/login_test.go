//go:build e2e

package e2e_test

import (
	"regexp"
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
	require.NoError(t, p.Locator(`form button[type="submit"]`).Click())

	require.NoError(t, expect.Page(p.Page).ToHaveURL(regexp.MustCompile(`/dashboard$`)))
}

func TestAUserLogsInThroughTheBrowserAndLogsOutFromTheUserMenu(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	ada := running.createAda(t)
	page := running.open(t, route.Login)

	page.logIn(t, ada)

	require.NoError(t, expect.Locator(page.Locator("#user-menu-name")).ToHaveText(ada.Name))

	require.NoError(t, page.Locator("#user-menu-name").Click())
	require.NoError(t, page.Locator(`form[action="/logout"] button`).Click())

	require.NoError(t, expect.Page(page.Page).ToHaveURL(regexp.MustCompile(`/$`)))
	require.NoError(t, expect.Locator(page.Locator("main h1")).ToHaveText("Let's get started"))
	assert.Empty(t, page.errors.list())
}
