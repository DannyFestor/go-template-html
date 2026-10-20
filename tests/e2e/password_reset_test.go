//go:build e2e

package e2e_test

import (
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/testsupport"
)

const newPassword = "a brand new password"

func TestAUserResetsTheirForgottenPasswordThroughTheBrowserAndLogsIn(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	ada := testsupport.Must(running.factory().User().WithEmail("ada@example.com").Create(t.Context()))(t)
	page := running.open(t, "/login")

	require.NoError(t, page.Locator(`a[href="/forgot-password"]`).Click())
	require.NoError(t, page.Locator("#email").Fill(ada.Email))
	require.NoError(t, page.Locator(`form button[type="submit"]`).Click())
	require.NoError(t, expect.Locator(page.Locator(`[role="status"]`)).
		ToHaveText("If an account exists, a reset link has been sent."))

	_, err := page.Goto(running.server.URL + running.mailedResetLink(t).RequestURI())
	require.NoError(t, err)
	require.NoError(t, page.Locator("#password").Fill(newPassword))
	require.NoError(t, page.Locator("#password_confirmation").Fill(newPassword))
	require.NoError(t, page.Locator(`form button[type="submit"]`).Click())

	require.NoError(t, expect.Page(page.Page).ToHaveURL(regexp.MustCompile(`/login$`)))
	require.NoError(t, expect.Locator(page.Locator(`[role="status"]`)).
		ToHaveText("Your password has been reset. Log in with your new password."))

	require.NoError(t, page.Locator("#email").Fill(ada.Email))
	require.NoError(t, page.Locator("#password").Fill(newPassword))
	require.NoError(t, page.Locator(`form button[type="submit"]`).Click())

	require.NoError(t, expect.Page(page.Page).ToHaveURL(regexp.MustCompile(`/dashboard$`)))
	assert.Empty(t, page.errors.list())
}

func (a *app) mailedResetLink(t *testing.T) *url.URL {
	t.Helper()

	a.RunQueuedJobs(t)

	sent := a.Outbox.Sent()
	require.Len(t, sent, 1)

	found := regexp.MustCompile(`Reset your password: (\S+)`).FindStringSubmatch(sent[0].Text)
	require.Len(t, found, 2, sent[0].Text)

	link, err := url.Parse(found[1])
	require.NoError(t, err)

	return link
}
