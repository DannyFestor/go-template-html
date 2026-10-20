//go:build feature

package feature_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/factory"
	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testapp"
)

const (
	incorrectPassword             = "The provided password is incorrect."
	confirmationOutlivedInstantly = time.Microsecond
)

func (a *app) signInAsAda(t *testing.T) {
	t.Helper()

	a.createAda(t)
	require.Equal(t, http.StatusSeeOther, a.login(t, credentials(adaEmail, factory.Password)).status)
}

func (a *app) confirmPassword(t *testing.T, password string) response {
	t.Helper()

	return a.post(t, route.ConfirmPassword, url.Values{"password": {password}})
}

func startAppWhosePasswordConfirmationsExpireAtOnce(t *testing.T) *app {
	t.Helper()

	config := liveConfig(t)
	config.Session.PasswordConfirmationTimeout = confirmationOutlivedInstantly

	return startAppOn(t, config)
}

func assertSentToConfirmThePassword(t *testing.T, running *app) {
	t.Helper()

	fetched := running.get(t, probeSensitivePage)

	assert.Equal(t, http.StatusSeeOther, fetched.status)
	assert.Equal(t, route.ConfirmPassword, fetched.header.Get("Location"))
}

func TestASensitivePageSendsTheUserToConfirmTheirPassword(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.signInAsAda(t)

	assertSentToConfirmThePassword(t, running)
}

func TestTheConfirmPasswordPageShowsThePasswordForm(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.signInAsAda(t)

	form := running.page(t, route.ConfirmPassword).
		Find(`main#public-main form[method="post"][action="/confirm-password"]`)

	require.Equal(t, 1, form.Length())
	assert.Equal(t, "false", form.AttrOr("hx-boost", ""))
	assert.Equal(t, "current-password", form.Find(`input[name="password"]`).AttrOr("autocomplete", ""))
}

func TestConfirmingThePasswordSendsTheUserBackToTheSensitivePage(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.signInAsAda(t)
	assertSentToConfirmThePassword(t, running)

	confirmed := running.confirmPassword(t, factory.Password)

	require.Equal(t, http.StatusSeeOther, confirmed.status)
	require.Equal(t, probeSensitivePage, confirmed.header.Get("Location"))

	fetched := running.get(t, probeSensitivePage)
	assert.Equal(t, http.StatusOK, fetched.status)
	assert.Equal(t, sensitivePageContent, string(fetched.body))
}

func TestConfirmingThePasswordWithoutASensitivePageAskedForLandsOnTheDashboard(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.signInAsAda(t)

	assert.Equal(t, route.Dashboard, running.confirmPassword(t, factory.Password).header.Get("Location"))
}

func TestAWrongOrMissingPasswordIsNotConfirmed(t *testing.T) {
	t.Parallel()

	for password, message := range map[string]string{
		"not the password": incorrectPassword,
		"":                 "Enter your password.",
	} {
		t.Run(message, func(t *testing.T) {
			t.Parallel()

			running := startApp(t)
			running.signInAsAda(t)

			refused := running.confirmPassword(t, password)

			assert.Equal(t, http.StatusUnprocessableEntity, refused.status)
			document := refused.document(t)
			assert.Equal(t, message, document.Find("#password-error").Text())
			assert.Empty(t, document.Find(`input[name="password"]`).AttrOr("value", ""))
			assertSentToConfirmThePassword(t, running)
		})
	}
}

func TestAPasswordConfirmationExpires(t *testing.T) {
	t.Parallel()

	running := startAppWhosePasswordConfirmationsExpireAtOnce(t)
	running.signInAsAda(t)
	require.Equal(t, http.StatusSeeOther, running.confirmPassword(t, factory.Password).status)

	assertSentToConfirmThePassword(t, running)
}

func TestPasswordConfirmationIsRateLimitedPerUser(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.signInAsAda(t)

	for range testapp.RateLimitAttempts {
		require.Equal(t, http.StatusUnprocessableEntity, running.confirmPassword(t, "wrong").status)
	}

	refused := running.confirmPassword(t, factory.Password)

	assert.Equal(t, http.StatusTooManyRequests, refused.status)
	assert.NotEmpty(t, refused.header.Get("Retry-After"))
	assert.Equal(t, rateLimitedMessage, refused.document(t).Find(`form [role="alert"]`).Text())
	assertSentToConfirmThePassword(t, running)
}

func TestAGuestIsSentToLoginInsteadOfConfirmingAPassword(t *testing.T) {
	t.Parallel()

	running := startApp(t)

	for _, path := range []string{probeSensitivePage, route.ConfirmPassword} {
		fetched := running.get(t, path)

		assert.Equal(t, http.StatusSeeOther, fetched.status, path)
		assert.Equal(t, route.Login, fetched.header.Get("Location"), path)
	}
}
