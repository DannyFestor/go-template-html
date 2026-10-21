//go:build feature

package feature_test

import (
	"net/http"
	"net/url"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/factory"
	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/testsupport"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testapp"
)

const (
	verificationSubject   = "Verify Email Address"
	verificationLinkSent  = "A new verification link has been sent to the email address you provided during registration."
	verificationExpiresAt = time.Nanosecond
)

func (a *app) registerAda(t *testing.T) {
	t.Helper()

	require.Equal(t, http.StatusSeeOther, a.register(t, registration(adaEmail, "correct horse")).status)
}

func (a *app) createUnverifiedAda(t *testing.T) {
	t.Helper()

	testsupport.Must(a.factory().User().WithEmail(adaEmail).Unverified().Create(t.Context()))(t)
}

func (a *app) loginAs(t *testing.T, email string) {
	t.Helper()

	require.Equal(t, http.StatusSeeOther, a.login(t, credentials(email, factory.Password)).status)
}

func (a *app) resendVerification(t *testing.T) response {
	t.Helper()

	return a.post(t, route.VerifyEmail, url.Values{})
}

func (a *app) verificationMailsToAda(t *testing.T) []domain.MailMessage {
	t.Helper()

	a.RunQueuedJobs(t)

	var verifications []domain.MailMessage

	for _, sent := range a.Outbox.Sent() {
		if sent.To == adaEmail && sent.Subject == verificationSubject {
			verifications = append(verifications, sent)
		}
	}

	return verifications
}

func (a *app) adasVerificationLink(t *testing.T) *url.URL {
	t.Helper()

	mails := a.verificationMailsToAda(t)
	require.NotEmpty(t, mails, "no verification mail was sent")
	latest := mails[len(mails)-1]

	html, err := goquery.NewDocumentFromReader(strings.NewReader(latest.HTML))
	require.NoError(t, err)

	link, err := url.Parse(html.Find("a").AttrOr("href", ""))
	require.NoError(t, err)
	require.Contains(t, latest.Text, link.String())

	return link
}

func startAppWhoseVerificationLinksExpireAtOnce(t *testing.T) *app {
	t.Helper()

	config := liveConfig(t)
	config.SignedURL.EmailVerificationTTL = verificationExpiresAt

	return startAppOn(t, config)
}

func assertSentToTheVerificationNotice(t *testing.T, running *app) {
	t.Helper()

	dashboard := running.get(t, route.Dashboard)

	assert.Equal(t, http.StatusSeeOther, dashboard.status)
	assert.Equal(t, route.VerifyEmail, dashboard.header.Get("Location"))
}

func TestARegisteredUserIsSentToTheVerificationNotice(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.registerAda(t)

	assertSentToTheVerificationNotice(t, running)

	notice := running.page(t, route.VerifyEmail)
	assert.Equal(t, 1, notice.Find(`main#public-main form[method="post"][action="/verify-email"] button`).Length())
	assert.Equal(t, 1, notice.Find(`main#public-main form[method="post"][action="/logout"] button`).Length())
}

func TestRegisteringMailsTheVerificationLink(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.registerAda(t)

	link := running.adasVerificationLink(t)

	assert.Regexp(t, `^/verify-email/[0-9a-f-]{36}/[0-9a-f]{64}$`, link.Path)
	assert.NotEmpty(t, link.Query().Get("signature"))
	assert.NotEmpty(t, link.Query().Get("expires"))
}

func TestFollowingTheVerificationLinkVerifiesTheUser(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.registerAda(t)
	assertSentToTheVerificationNotice(t, running)

	verified := running.get(t, running.adasVerificationLink(t).RequestURI())

	require.Equal(t, http.StatusSeeOther, verified.status)
	assert.Equal(t, route.Dashboard, verified.header.Get("Location"))

	dashboard := running.page(t, route.Dashboard)
	assert.Equal(t, "Your email address has been verified.", dashboard.Find(`main [role="status"]`).Text())
	assert.Equal(t, "Ada Lovelace", dashboard.Find("#user-menu-name").Text())
}

func TestAGuestFollowingTheVerificationLinkIsVerifiedAfterLoggingIn(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.registerAda(t)
	link := running.adasVerificationLink(t).RequestURI()
	require.Equal(t, http.StatusSeeOther, running.post(t, route.Logout, url.Values{}).status)

	require.Equal(t, route.Login, running.get(t, link).header.Get("Location"))
	loggedIn := running.login(t, credentials(adaEmail, "correct horse"))
	require.Equal(t, link, loggedIn.header.Get("Location"))

	require.Equal(t, http.StatusSeeOther, running.get(t, link).status)
	assert.Equal(t, http.StatusOK, running.get(t, route.Dashboard).status)
}

func TestATamperedVerificationLinkIsForbidden(t *testing.T) {
	t.Parallel()

	tamperings := map[string]func(*url.URL){
		"hash":              func(link *url.URL) { link.Path = path.Dir(link.Path) + "/" + strings.Repeat("0", 64) },
		"extended expiry":   func(link *url.URL) { setQueryValue(link, "expires", "99999999999") },
		"forged signature":  func(link *url.URL) { setQueryValue(link, "signature", strings.Repeat("0", 64)) },
		"removed signature": func(link *url.URL) { link.RawQuery = "" },
	}

	for name, tamper := range tamperings {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			running := startApp(t)
			running.registerAda(t)
			link := running.adasVerificationLink(t)
			tamper(link)

			assert.Equal(t, http.StatusForbidden, running.get(t, link.RequestURI()).status)
			assertSentToTheVerificationNotice(t, running)
		})
	}
}

func TestAnExpiredVerificationLinkIsForbidden(t *testing.T) {
	t.Parallel()

	running := startAppWhoseVerificationLinksExpireAtOnce(t)
	running.registerAda(t)

	assert.Equal(t, http.StatusForbidden, running.get(t, running.adasVerificationLink(t).RequestURI()).status)
	assertSentToTheVerificationNotice(t, running)
}

func TestAVerificationLinkIssuedToAnotherUserIsForbidden(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.registerAda(t)
	adasLink := running.adasVerificationLink(t).RequestURI()
	require.Equal(t, http.StatusSeeOther, running.post(t, route.Logout, url.Values{}).status)
	testsupport.Must(running.factory().User().WithEmail("grace@example.com").Unverified().Create(t.Context()))(t)
	running.loginAs(t, "grace@example.com")

	assert.Equal(t, http.StatusForbidden, running.get(t, adasLink).status)
	assertSentToTheVerificationNotice(t, running)
}

func TestResendingMailsAFreshVerificationLink(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.createUnverifiedAda(t)
	running.loginAs(t, adaEmail)

	resent := running.resendVerification(t)

	require.Equal(t, http.StatusSeeOther, resent.status)
	assert.Equal(t, route.VerifyEmail, resent.header.Get("Location"))
	assert.Equal(t, verificationLinkSent, running.page(t, route.VerifyEmail).Find(`main [role="status"]`).Text())
	assert.Len(t, running.verificationMailsToAda(t), 1)
}

func TestResendingTheVerificationMailIsRateLimitedPerUser(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.createUnverifiedAda(t)
	running.loginAs(t, adaEmail)

	for range testapp.RateLimitAttempts {
		require.Equal(t, http.StatusSeeOther, running.resendVerification(t).status)
	}

	refused := running.resendVerification(t)

	assert.Equal(t, http.StatusTooManyRequests, refused.status)
	assert.NotEmpty(t, refused.header.Get("Retry-After"))
	assert.Equal(t, rateLimitedMessage, refused.document(t).Find(`form [role="alert"]`).Text())
	assert.Len(t, running.verificationMailsToAda(t), testapp.RateLimitAttempts)
}

func TestFollowingTheVerificationLinkIsRateLimitedPerUser(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.registerAda(t)
	link := running.adasVerificationLink(t).RequestURI()

	for range testapp.RateLimitAttempts {
		require.Equal(t, http.StatusSeeOther, running.get(t, link).status)
	}

	refused := running.get(t, link)

	assert.Equal(t, http.StatusTooManyRequests, refused.status)
	assert.NotEmpty(t, refused.header.Get("Retry-After"))
}

func TestAVerifiedUserIsSentFromTheVerificationNoticeToTheDashboard(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	running.createAda(t)
	running.loginAs(t, adaEmail)

	for name, fetched := range map[string]response{
		"notice": running.get(t, route.VerifyEmail),
		"resend": running.resendVerification(t),
	} {
		assert.Equal(t, http.StatusSeeOther, fetched.status, name)
		assert.Equal(t, route.Dashboard, fetched.header.Get("Location"), name)
	}

	assert.Empty(t, running.verificationMailsToAda(t))
}

func TestAGuestRequestingTheVerificationNoticeIsSentToLogin(t *testing.T) {
	t.Parallel()

	fetched := startApp(t).get(t, route.VerifyEmail)

	assert.Equal(t, http.StatusSeeOther, fetched.status)
	assert.Equal(t, route.Login, fetched.header.Get("Location"))
}

func setQueryValue(link *url.URL, name, value string) {
	query := link.Query()
	query.Set(name, value)
	link.RawQuery = query.Encode()
}
