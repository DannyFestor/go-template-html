//go:build feature

package feature_test

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/bootstrap"
	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/factory"
	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testapp"
)

const (
	resetLinkSent        = "If an account exists, a reset link has been sent."
	passwordWasReset     = "Your password has been reset. Log in with your new password."
	invalidResetLink     = "This password reset link is invalid or has expired."
	newPassword          = "a brand new password"
	resetLinkOutlivedNow = time.Microsecond
)

func (a *app) requestPasswordReset(t *testing.T, email string) response {
	t.Helper()

	return a.post(t, route.ForgotPassword, url.Values{"email": {email}})
}

func (a *app) resetPassword(t *testing.T, token, password string) response {
	t.Helper()

	return a.post(t, route.ResetPassword, url.Values{
		"token":                 {token},
		"password":              {password},
		"password_confirmation": {password},
	})
}

func (a *app) mailedResetLinks(t *testing.T) []*url.URL {
	t.Helper()

	a.RunQueuedJobs(t)

	sent := a.Outbox.Sent()
	links := make([]*url.URL, 0, len(sent))

	for _, sent := range sent {
		links = append(links, resetLinkIn(t, sent))
	}

	return links
}

func (a *app) adasMailedResetToken(t *testing.T) string {
	t.Helper()

	require.Equal(t, http.StatusSeeOther, a.requestPasswordReset(t, adaEmail).status)

	links := a.mailedResetLinks(t)
	require.NotEmpty(t, links)

	return links[len(links)-1].Query().Get("token")
}

func resetLinkIn(t *testing.T, sent domain.MailMessage) *url.URL {
	t.Helper()

	html, err := goquery.NewDocumentFromReader(strings.NewReader(sent.HTML))
	require.NoError(t, err)

	link, err := url.Parse(html.Find("a[href]").AttrOr("href", ""))
	require.NoError(t, err)

	return link
}

func startAppWhoseResetLinksExpireAtOnce(t *testing.T) *app {
	t.Helper()

	config := liveConfig(t)
	config.PasswordReset.TTL = resetLinkOutlivedNow

	return startAppOn(t, config)
}

func startAppAllowingTwoResetRequestsPerEmail(t *testing.T) *app {
	t.Helper()

	config := liveConfig(t)
	config.RateLimits.ForgotPasswordEmail = bootstrap.RateLimit{MaxAttempts: 2, Window: time.Minute}

	return startAppOn(t, config)
}

func assertResetRefused(t *testing.T, refused response) {
	t.Helper()

	assert.Equal(t, http.StatusUnprocessableEntity, refused.status)
	assert.Equal(t, invalidResetLink, refused.document(t).Find(`form [role="alert"]`).Text())
}

func assertLogsInWith(t *testing.T, running *app, password string) {
	t.Helper()

	loggedIn := running.login(t, credentials(adaEmail, password))

	assert.Equal(t, http.StatusSeeOther, loggedIn.status)
	assert.Equal(t, route.Dashboard, loggedIn.header.Get("Location"))
}

func TestTheForgotPasswordPageShowsTheForm(t *testing.T) {
	t.Parallel()

	document := startApp(t).page(t, route.ForgotPassword)

	form := document.Find(`main#public-main form[method="post"][action="/forgot-password"]`)
	require.Equal(t, 1, form.Length())
	assert.Equal(t, "false", form.AttrOr("hx-boost", ""))
	assert.Equal(t, 1, form.Find(`input[name="email"]`).Length())
	assert.Equal(t, 1, document.Find(`main#public-main a[href="/login"]`).Length())
}

func TestAKnownAndAnUnknownEmailGetTheSameResponse(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.createAda(t)

	known := running.requestPasswordReset(t, adaEmail)
	knownStatus := running.page(t, route.ForgotPassword).Find(`[role="status"]`).Text()
	unknown := running.requestPasswordReset(t, "grace@example.com")
	unknownStatus := running.page(t, route.ForgotPassword).Find(`[role="status"]`).Text()

	assert.Equal(t, http.StatusSeeOther, known.status)
	assert.Equal(t, known.status, unknown.status)
	assert.Equal(t, route.ForgotPassword, known.header.Get("Location"))
	assert.Equal(t, known.header.Get("Location"), unknown.header.Get("Location"))
	assert.Equal(t, known.body, unknown.body)
	assert.Equal(t, resetLinkSent, knownStatus)
	assert.Equal(t, knownStatus, unknownStatus)

	running.RunQueuedJobs(t)

	sent := running.Outbox.Sent()
	require.Len(t, sent, 1)
	assert.Equal(t, adaEmail, sent[0].To)
}

func TestTheResetMailLinksToTheResetPageInBothParts(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.createAda(t)
	require.Equal(t, http.StatusSeeOther, running.requestPasswordReset(t, "ADA@example.com").status)

	running.RunQueuedJobs(t)

	sent := running.Outbox.Sent()
	require.Len(t, sent, 1)
	link := resetLinkIn(t, sent[0])
	assert.Equal(t, "Reset your password", sent[0].Subject)
	assert.Equal(t, "http://127.0.0.1"+route.ResetPassword, link.Scheme+"://"+link.Host+link.Path)
	assert.Regexp(t, `^[A-Z2-7]{26}$`, link.Query().Get("token"))
	assert.Contains(t, sent[0].Text, "Reset your password: "+link.String()+"\n")
	assert.Contains(t, sent[0].Text, "expire in 60 minutes")
}

func TestAMalformedEmailIsRefusedWithAFieldError(t *testing.T) {
	t.Parallel()

	refused := startApp(t).requestPasswordReset(t, "ada@")

	assert.Equal(t, http.StatusUnprocessableEntity, refused.status)
	document := refused.document(t)
	assert.Equal(t, "Enter a valid email address.", document.Find("#email-error").Text())
	assert.Equal(t, "ada@", document.Find(`input[name="email"]`).AttrOr("value", ""))
}

func TestTheResetPageCarriesTheTokenWithoutLeakingIt(t *testing.T) {
	t.Parallel()

	fetched := startApp(t).send(t, acceptingEncoding(route.ResetPassword+"?token=RESETTOKEN", "gzip, zstd"))

	assert.Equal(t, http.StatusOK, fetched.status)
	assert.Equal(t, "no-referrer", fetched.header.Get("Referrer-Policy"))
	assert.Empty(t, fetched.header.Get("Content-Encoding"))

	form := fetched.document(t).Find(`main#public-main form[method="post"][action="/reset-password"]`)
	require.Equal(t, 1, form.Length())
	assert.Equal(t, "RESETTOKEN", form.Find(`input[type="hidden"][name="token"]`).AttrOr("value", ""))

	for _, name := range []string{"password", "password_confirmation"} {
		assert.Equal(t, 1, form.Find(`input[name="`+name+`"]`).Length(), name)
	}
}

func TestAValidTokenResetsThePasswordAndTheUserMustThenLogIn(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.createAda(t)
	token := running.adasMailedResetToken(t)

	reset := running.resetPassword(t, token, newPassword)

	require.Equal(t, http.StatusSeeOther, reset.status)
	require.Equal(t, route.Login, reset.header.Get("Location"))
	assert.Equal(t, passwordWasReset, running.page(t, route.Login).Find(`main#public-main [role="status"]`).Text())
	assertStillAGuest(t, running)
	assert.Equal(t, http.StatusUnprocessableEntity, running.login(t, credentials(adaEmail, factory.Password)).status)
	assertLogsInWith(t, running, newPassword)
}

func TestAReusedTokenFails(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.createAda(t)
	token := running.adasMailedResetToken(t)
	require.Equal(t, http.StatusSeeOther, running.resetPassword(t, token, newPassword).status)

	assertResetRefused(t, running.resetPassword(t, token, "yet another password"))
	assertLogsInWith(t, running, newPassword)
}

func TestAnExpiredTokenFails(t *testing.T) {
	t.Parallel()

	running := startAppWhoseResetLinksExpireAtOnce(t)
	running.createAda(t)
	token := running.adasMailedResetToken(t)

	assertResetRefused(t, running.resetPassword(t, token, newPassword))
	assertLogsInWith(t, running, factory.Password)
}

func TestAnUnknownTokenFailsAndKeepsTheFormFilledWithIt(t *testing.T) {
	t.Parallel()

	refused := startApp(t).resetPassword(t, "UNKNOWNTOKEN", newPassword)

	assertResetRefused(t, refused)
	assert.Equal(t, "UNKNOWNTOKEN", refused.document(t).Find(`input[name="token"]`).AttrOr("value", ""))
}

func TestANewRequestReplacesTheOldToken(t *testing.T) {
	t.Parallel()

	running := startAppAllowingTwoResetRequestsPerEmail(t)
	running.createAda(t)
	first := running.adasMailedResetToken(t)
	second := running.adasMailedResetToken(t)

	assertResetRefused(t, running.resetPassword(t, first, newPassword))
	assert.Equal(t, http.StatusSeeOther, running.resetPassword(t, second, newPassword).status)
}

func TestAResetWithAnInvalidPasswordReRendersWithoutUsingUpTheToken(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.createAda(t)
	token := running.adasMailedResetToken(t)

	refused := running.post(t, route.ResetPassword, url.Values{
		"token":                 {token},
		"password":              {"short"},
		"password_confirmation": {"other"},
	})

	assert.Equal(t, http.StatusUnprocessableEntity, refused.status)
	document := refused.document(t)
	assert.Equal(t, "The password must be at least 8 characters.", document.Find("#password-error").Text())
	assert.Equal(t, token, document.Find(`input[name="token"]`).AttrOr("value", ""))
	assert.Equal(t, http.StatusSeeOther, running.resetPassword(t, token, newPassword).status)
}

func TestAPasswordResetSignsOutOtherSessionsAndForgetsTheRememberedDevices(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.createAda(t)
	require.Equal(t, http.StatusSeeOther, running.login(t, adasRememberedCredentials()).status)
	remembered := running.rememberCookie(t)
	require.NotEmpty(t, remembered)

	fromAnotherDevice := func(path string, form url.Values) response {
		return running.sendWithoutCookieJar(t, outgoing{
			method:  http.MethodPost,
			path:    path,
			headers: map[string]string{"Content-Type": "application/x-www-form-urlencoded"},
			body:    strings.NewReader(form.Encode()),
		})
	}
	require.Equal(
		t,
		http.StatusSeeOther,
		fromAnotherDevice(route.ForgotPassword, url.Values{"email": {adaEmail}}).status,
	)
	token := running.mailedResetLinks(t)[0].Query().Get("token")
	require.Equal(t, http.StatusSeeOther, fromAnotherDevice(route.ResetPassword, url.Values{
		"token":                 {token},
		"password":              {newPassword},
		"password_confirmation": {newPassword},
	}).status)

	assertStillAGuest(t, running)
	running.replayRememberCookie(t, remembered)
	assertStillAGuest(t, running)
}

func TestTheResetRequestSendsOneMailPerEmailPerWindowAndLooksTheSameAfter(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.createAda(t)
	first := running.requestPasswordReset(t, adaEmail)
	running.page(t, route.ForgotPassword)

	second := running.requestPasswordReset(t, "ADA@example.com")

	assert.Equal(t, first.status, second.status)
	assert.Equal(t, first.header.Get("Location"), second.header.Get("Location"))
	assert.Equal(t, resetLinkSent, running.page(t, route.ForgotPassword).Find(`[role="status"]`).Text())
	assert.Len(t, running.mailedResetLinks(t), 1)
}

func TestTheResetRequestIsRateLimitedPerClientIP(t *testing.T) {
	t.Parallel()

	running := startApp(t)

	for attempt := range testapp.RateLimitAttempts {
		require.Equal(t, http.StatusSeeOther,
			running.requestPasswordReset(t, fmt.Sprintf("visitor%d@example.com", attempt)).status)
	}

	refused := running.requestPasswordReset(t, "another@example.com")

	assert.Equal(t, http.StatusTooManyRequests, refused.status)
	assert.NotEmpty(t, refused.header.Get("Retry-After"))
	assert.Equal(t, rateLimitedMessage, refused.document(t).Find(`form [role="alert"]`).Text())
}

func TestTheResetSubmitIsRateLimitedPerClientIP(t *testing.T) {
	t.Parallel()

	running := startApp(t)

	for range testapp.RateLimitAttempts {
		require.Equal(t, http.StatusUnprocessableEntity, running.resetPassword(t, "UNKNOWNTOKEN", newPassword).status)
	}

	refused := running.resetPassword(t, "UNKNOWNTOKEN", newPassword)

	assert.Equal(t, http.StatusTooManyRequests, refused.status)
	assert.NotEmpty(t, refused.header.Get("Retry-After"))
	assert.Equal(t, rateLimitedMessage, refused.document(t).Find(`form [role="alert"]`).Text())
}

func TestASignedInUserRequestingThePasswordResetPagesIsSentToTheDashboard(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.createAda(t)
	require.Equal(t, http.StatusSeeOther, running.login(t, credentials(adaEmail, factory.Password)).status)

	for _, path := range []string{route.ForgotPassword, route.ResetPassword + "?token=RESETTOKEN"} {
		fetched := running.get(t, path)

		assert.Equal(t, http.StatusSeeOther, fetched.status, path)
		assert.Equal(t, route.Dashboard, fetched.header.Get("Location"), path)
	}
}
