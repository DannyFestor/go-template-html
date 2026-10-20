//go:build e2e

package e2e_test

import (
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAVisitorRegistersVerifiesTheirEmailAndLandsOnTheDashboard(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	page := running.open(t, "/register")

	for field, value := range map[string]string{
		"#name":                  "Ada Lovelace",
		"#email":                 "ada@example.com",
		"#password":              "correct horse",
		"#password_confirmation": "correct horse",
	} {
		require.NoError(t, page.Locator(field).Fill(value), field)
	}

	require.NoError(t, page.Locator(`form button[type="submit"]`).Click())

	require.NoError(t, expect.Page(page.Page).ToHaveURL(regexp.MustCompile(`/verify-email$`)))
	require.NoError(t, expect.Locator(page.Locator("main h1")).ToHaveText("Verify your email address"))

	_, err := page.Goto(running.server.URL + running.verificationLinkSentTo(t, "ada@example.com"))
	require.NoError(t, err)

	require.NoError(t, expect.Page(page.Page).ToHaveURL(regexp.MustCompile(`/dashboard$`)))
	require.NoError(t, expect.Locator(page.Locator(`main [role="status"]`)).
		ToHaveText("Your email address has been verified."))
	require.NoError(t, expect.Locator(page.Locator("#user-menu-name")).ToHaveText("Ada Lovelace"))
	assert.Empty(t, page.errors.list())
}

func (a *app) verificationLinkSentTo(t *testing.T, recipient string) string {
	t.Helper()

	a.RunQueuedJobs(t)

	sent := a.Outbox.Sent()
	require.Len(t, sent, 1)
	require.Equal(t, recipient, sent[0].To)

	link, err := url.Parse(regexp.MustCompile(`https?://\S+`).FindString(sent[0].Text))
	require.NoError(t, err)

	return link.RequestURI()
}
