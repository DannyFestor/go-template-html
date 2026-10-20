//go:build feature

package feature_test

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/factory"
	authjob "github.com/DannyFestor/go-template-html/internal/job/auth"
	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/testsupport"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testapp"
)

const (
	adaEmail                 = "ada@example.com"
	noMatchingAccount        = "These credentials do not match our records."
	rateLimitedMessage       = "Too many attempts. Please try again later."
	rememberCookieName       = "remember"
	sessionOutlivedInstantly = time.Microsecond
)

func credentials(email, password string) url.Values {
	return url.Values{"email": {email}, "password": {password}}
}

func adasRememberedCredentials() url.Values {
	form := credentials(adaEmail, factory.Password)
	form.Set("remember", "on")

	return form
}

func (a *app) login(t *testing.T, form url.Values) response {
	t.Helper()

	return a.post(t, route.Login, form)
}

func (a *app) createAda(t *testing.T) domain.User {
	t.Helper()

	return testsupport.Must(a.factory().User().WithEmail(adaEmail).Create(t.Context()))(t)
}

func (a *app) rememberCookie(t *testing.T) string {
	t.Helper()

	for _, cookie := range a.cookies(t) {
		if cookie.Name == rememberCookieName {
			return cookie.Value
		}
	}

	return ""
}

func (a *app) replayRememberCookie(t *testing.T, value string) {
	t.Helper()

	replayed, err := http.ParseSetCookie(rememberCookieName + "=" + value + "; Path=/; HttpOnly; SameSite=Lax")
	require.NoError(t, err)

	a.server.Client().Jar.SetCookies(a.address(t), []*http.Cookie{replayed})
}

func startAppWhoseSessionsExpireAtOnce(t *testing.T) *app {
	t.Helper()

	config := liveConfig(t)
	config.Session.AbsoluteLifetime = sessionOutlivedInstantly

	return startAppOn(t, config)
}

func TestTheLoginPageShowsTheLoginForm(t *testing.T) {
	t.Parallel()

	document := startApp(t).page(t, route.Login)

	form := document.Find(`main#public-main form[method="post"][action="/login"]`)
	require.Equal(t, 1, form.Length())
	assert.Equal(t, "false", form.AttrOr("hx-boost", ""))

	for _, field := range []string{`input[name="email"]`, `input[name="password"]`, `input[name="remember"]`} {
		assert.Equal(t, 1, form.Find(field).Length(), field)
	}

	for _, link := range []string{route.Register, route.ForgotPassword} {
		assert.Equal(t, 1, document.Find(`main#public-main a[href="`+link+`"]`).Length(), link)
	}
}

func TestAUserLogsInWithValidCredentialsAndLandsOnTheDashboard(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	ada := running.createAda(t)

	loggedIn := running.login(t, credentials("ADA@example.com", factory.Password))

	require.Equal(t, http.StatusSeeOther, loggedIn.status)
	require.Equal(t, route.Dashboard, loggedIn.header.Get("Location"))
	assert.Equal(t, ada.Name, running.page(t, route.Dashboard).Find("#user-menu-name").Text())
}

func TestALoginWithWrongCredentialsReRendersWithoutRevealingWhichFieldFailed(t *testing.T) {
	t.Parallel()

	for name, attempt := range map[string]url.Values{
		"wrong password": credentials(adaEmail, "not the password"),
		"unknown email":  credentials("grace@example.com", factory.Password),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			running := startApp(t)
			running.createAda(t)

			refused := running.login(t, attempt)

			assert.Equal(t, http.StatusUnprocessableEntity, refused.status)
			document := refused.document(t)
			assert.Equal(t, noMatchingAccount, document.Find(`form [role="alert"]`).Text())
			assert.Zero(t, document.Find("#email-error, #password-error").Length())
			assert.Equal(t, attempt.Get("email"), document.Find(`input[name="email"]`).AttrOr("value", ""))
			assert.Empty(t, document.Find(`input[name="password"]`).AttrOr("value", ""))
			assertStillAGuest(t, running)
		})
	}
}

func TestLoginIsRateLimitedPerEmailAndClientIP(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.createAda(t)

	for range testapp.RateLimitAttempts {
		require.Equal(t, http.StatusUnprocessableEntity, running.login(t, credentials(adaEmail, "wrong")).status)
	}

	refused := running.login(t, credentials("ADA@example.com", factory.Password))

	assert.Equal(t, http.StatusTooManyRequests, refused.status)
	assert.NotEmpty(t, refused.header.Get("Retry-After"))
	assert.Equal(t, rateLimitedMessage, refused.document(t).Find(`form [role="alert"]`).Text())
	assert.Equal(t, http.StatusUnprocessableEntity, running.login(t, credentials("grace@example.com", "wrong")).status)
}

func TestASuccessfulLoginResetsItsRateLimit(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.createAda(t)

	for range testapp.RateLimitAttempts - 1 {
		require.Equal(t, http.StatusUnprocessableEntity, running.login(t, credentials(adaEmail, "wrong")).status)
	}

	require.Equal(t, http.StatusSeeOther, running.login(t, credentials(adaEmail, factory.Password)).status)
	require.Equal(t, http.StatusSeeOther, running.post(t, route.Logout, url.Values{}).status)

	for range testapp.RateLimitAttempts {
		assert.Equal(t, http.StatusUnprocessableEntity, running.login(t, credentials(adaEmail, "wrong")).status)
	}
}

func TestALoginSendsTheUserToThePageTheyAskedForAsAGuest(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.createAda(t)
	require.Equal(t, route.Login, running.get(t, route.Dashboard+"?tab=team").header.Get("Location"))

	loggedIn := running.login(t, credentials(adaEmail, factory.Password))

	assert.Equal(t, route.Dashboard+"?tab=team", loggedIn.header.Get("Location"))
}

func TestASignedInUserRequestingTheLoginPageIsSentToTheDashboard(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.createAda(t)
	require.Equal(t, http.StatusSeeOther, running.login(t, credentials(adaEmail, factory.Password)).status)

	fetched := running.get(t, route.Login)

	assert.Equal(t, http.StatusSeeOther, fetched.status)
	assert.Equal(t, route.Dashboard, fetched.header.Get("Location"))
}

func TestLoggingOutFromTheUserMenuEndsTheSession(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.createAda(t)
	require.Equal(t, http.StatusSeeOther, running.login(t, credentials(adaEmail, factory.Password)).status)

	menu := running.page(t, route.Dashboard).Find(`aside form[method="post"][action="/logout"]`)
	require.Equal(t, 1, menu.Length())

	loggedOut := running.post(t, route.Logout, url.Values{})

	assert.Equal(t, http.StatusSeeOther, loggedOut.status)
	assert.Equal(t, route.Home, loggedOut.header.Get("Location"))
	assertStillAGuest(t, running)
}

func TestARememberTokenRestoresTheSessionAfterTheSessionExpires(t *testing.T) {
	t.Parallel()

	running := startAppWhoseSessionsExpireAtOnce(t)
	ada := running.createAda(t)
	require.Equal(t, http.StatusSeeOther, running.login(t, adasRememberedCredentials()).status)
	issued := running.rememberCookie(t)

	assert.Equal(t, ada.Name, running.page(t, route.Dashboard).Find("#user-menu-name").Text())
	assert.NotEqual(t, issued, running.rememberCookie(t), "the remember token was not rotated")
}

func TestARequestRacingTheRestoreWithTheSameTokenIsSignedInAndKeepsTheRotatedCookie(t *testing.T) {
	t.Parallel()

	running := startAppWhoseSessionsExpireAtOnce(t)
	ada := running.createAda(t)
	require.Equal(t, http.StatusSeeOther, running.login(t, adasRememberedCredentials()).status)
	presented := running.rememberCookie(t)
	require.Equal(t, http.StatusOK, running.get(t, route.Dashboard).status)

	racing := running.sendWithoutCookieJar(t, outgoing{
		method:  http.MethodGet,
		path:    route.Dashboard,
		headers: map[string]string{"Cookie": rememberCookieName + "=" + presented},
		body:    nil,
	})

	require.Equal(t, http.StatusOK, racing.status)
	assert.Equal(t, ada.Name, racing.document(t).Find("#user-menu-name").Text())

	for _, setCookie := range racing.header.Values("Set-Cookie") {
		assert.NotContains(t, setCookie, rememberCookieName+"=", "the racing request touched the remember cookie")
	}
}

func TestWithoutRememberMeAnExpiredSessionSendsTheUserToLogin(t *testing.T) {
	t.Parallel()

	running := startAppWhoseSessionsExpireAtOnce(t)
	running.createAda(t)
	require.Equal(t, http.StatusSeeOther, running.login(t, credentials(adaEmail, factory.Password)).status)

	assertStillAGuest(t, running)
}

func TestLoggingOutForgetsTheRememberedDevice(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.createAda(t)
	require.Equal(t, http.StatusSeeOther, running.login(t, adasRememberedCredentials()).status)
	remembered := running.rememberCookie(t)
	require.NotEmpty(t, remembered)

	require.Equal(t, http.StatusSeeOther, running.post(t, route.Logout, url.Values{}).status)
	assert.Empty(t, running.rememberCookie(t))

	running.replayRememberCookie(t, remembered)
	assertStillAGuest(t, running)
}

func TestThePruneJobDeletesTheExpiredRememberTokens(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	ada := running.createAda(t)
	expired := domain.RememberToken{Selector: "EXPIREDSELECTOR", Validator: "EXPIREDVALIDATOR"}
	live := domain.RememberToken{Selector: "LIVESELECTOR", Validator: "LIVEVALIDATOR"}

	require.NoError(t, running.RememberTokens.Issue(t.Context(), ada.ID, expired, time.Now().Add(-time.Minute)))
	require.NoError(t, running.RememberTokens.Issue(t.Context(), ada.ID, live, time.Now().Add(time.Hour)))

	require.NoError(t, running.Jobs.Enqueue(t.Context(), authjob.PruneExpiredRememberTokens{}))
	running.RunQueuedJobs(t)

	assert.False(t, running.consumeRememberToken(t, expired), "expired")
	assert.True(t, running.consumeRememberToken(t, live), "live")
}

func (a *app) consumeRememberToken(t *testing.T, token domain.RememberToken) bool {
	t.Helper()

	var beforeAnyExpiry time.Time

	_, consumed, err := a.RememberTokens.Consume(t.Context(), token, beforeAnyExpiry)
	require.NoError(t, err)

	return consumed
}
