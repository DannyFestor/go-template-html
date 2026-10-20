package middleware_test

import (
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/middleware"
	"github.com/DannyFestor/go-template-html/internal/route"
)

const (
	confirmationTimeout = 3 * time.Hour
	sensitiveContent    = "Sensitive."
)

type confirmationApp struct {
	*sessionApp

	responder    *spyResponder
	confirmation *middleware.PasswordConfirmation
}

func newConfirmationApp() *confirmationApp {
	sessions := newSessionApp()
	responder := &spyResponder{rendered: nil, redirects: nil}

	return &confirmationApp{
		sessionApp:   sessions,
		responder:    responder,
		confirmation: middleware.NewPasswordConfirmation(sessions.clock, confirmationTimeout, responder),
	}
}

func (a *confirmationApp) signIn(t *testing.T) *http.Cookie {
	t.Helper()

	return a.serve(t, signingIn, nil).cookie(t)
}

func (a *confirmationApp) confirm(t *testing.T, cookie *http.Cookie) served {
	t.Helper()

	return a.serve(t, func(w http.ResponseWriter, r *http.Request) {
		a.confirmation.Confirm(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}, cookie)
}

func (a *confirmationApp) requestTheSensitivePage(t *testing.T, target string, cookie *http.Cookie) served {
	t.Helper()

	return a.serveAt(t, target, a.confirmation.Require(http.HandlerFunc(showingTheSensitivePage)).ServeHTTP, cookie)
}

func (a *confirmationApp) intendedAfterConfirming(t *testing.T, cookie *http.Cookie) string {
	t.Helper()

	return a.serve(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, a.confirmation.Intended(r.Context()))
	}, cookie).body
}

func showingTheSensitivePage(w http.ResponseWriter, _ *http.Request) {
	_, _ = io.WriteString(w, sensitiveContent)
}

func TestAnUnconfirmedUserIsSentToConfirmTheirPassword(t *testing.T) {
	t.Parallel()

	app := newConfirmationApp()

	guarded := app.requestTheSensitivePage(t, "/", app.signIn(t))

	assert.Equal(t, http.StatusSeeOther, guarded.status)
	assert.Equal(t, []string{route.ConfirmPassword}, app.responder.redirects)
	assert.Empty(t, guarded.body)
}

func TestTheSensitivePageIsIntendedOnceAfterConfirming(t *testing.T) {
	t.Parallel()

	app := newConfirmationApp()
	cookie := app.signIn(t)
	app.requestTheSensitivePage(t, "/settings/two-factor?step=enable", cookie)

	assert.Equal(t, "/settings/two-factor?step=enable", app.intendedAfterConfirming(t, cookie))
	assert.Equal(t, route.Dashboard, app.intendedAfterConfirming(t, cookie))
}

func TestConfirmingThePasswordReplacesTheSessionToken(t *testing.T) {
	t.Parallel()

	app := newConfirmationApp()
	cookie := app.signIn(t)

	confirmed := app.confirm(t, cookie).cookie(t)

	assert.NotEqual(t, cookie.Value, confirmed.Value)
	assert.Equal(t, "guest ", app.serve(t, showingTheUser, cookie).body)
	assert.Equal(t, userID+" "+fingerprint, app.serve(t, showingTheUser, confirmed).body)
}

func TestAConfirmationLetsTheUserThroughUntilItTimesOut(t *testing.T) {
	t.Parallel()

	app := newConfirmationApp()
	cookie := app.confirm(t, app.signIn(t)).cookie(t)

	app.clock.now = app.clock.now.Add(confirmationTimeout - time.Nanosecond)
	assert.Equal(t, sensitiveContent, app.requestTheSensitivePage(t, "/", cookie).body)

	app.clock.now = app.clock.now.Add(time.Nanosecond)
	assert.Equal(t, http.StatusSeeOther, app.requestTheSensitivePage(t, "/", cookie).status)
	assert.Equal(t, []string{route.ConfirmPassword}, app.responder.redirects)
}
