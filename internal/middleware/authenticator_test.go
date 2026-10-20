package middleware_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/middleware"
	"github.com/DannyFestor/go-template-html/internal/route"
)

var errUsersDown = errors.New("users down")

type fakeUsers struct {
	users map[domain.UserID]domain.User
	err   error
}

func (f *fakeUsers) Find(_ context.Context, id domain.UserID) (domain.User, bool, error) {
	user, found := f.users[id]

	return user, found, f.err
}

type labelledFingerprints struct{}

func (labelledFingerprints) Of(hash domain.PasswordHash) string {
	return "fingerprint of " + string(hash)
}

func (f labelledFingerprints) Matches(hash domain.PasswordHash, fingerprint string) bool {
	return f.Of(hash) == fingerprint
}

type spyResponder struct {
	spyRenderer

	redirects []string
}

func (s *spyResponder) Redirect(w http.ResponseWriter, _ *http.Request, url string) {
	s.redirects = append(s.redirects, url)

	w.WriteHeader(http.StatusSeeOther)
}

type authApp struct {
	*sessionApp

	users         *fakeUsers
	responder     *spyResponder
	authenticator *middleware.Authenticator
}

func newAuthApp() *authApp {
	users := &fakeUsers{users: map[domain.UserID]domain.User{userID: ada()}, err: nil}
	responder := &spyResponder{rendered: nil, redirects: nil}

	return &authApp{
		sessionApp:    newSessionApp(),
		users:         users,
		responder:     responder,
		authenticator: middleware.NewAuthenticator(users, labelledFingerprints{}, responder),
	}
}

func ada() domain.User {
	return domain.User{
		ID:           userID,
		Name:         "Ada Lovelace",
		Email:        "ada@example.com",
		PasswordHash: "hash-1",
		CreatedAt:    time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC),
	}
}

func (a *authApp) signIn(t *testing.T) *http.Cookie {
	t.Helper()

	return a.serve(t, func(w http.ResponseWriter, r *http.Request) {
		a.authenticator.SignIn(r.Context(), ada())
		w.WriteHeader(http.StatusNoContent)
	}, nil).cookie(t)
}

func (a *authApp) requestTheGuardedPage(t *testing.T, cookie *http.Cookie) served {
	t.Helper()

	return a.serve(t, a.authenticator.Require(http.HandlerFunc(greetingTheUser)).ServeHTTP, cookie)
}

func greetingTheUser(w http.ResponseWriter, r *http.Request) {
	user, signedIn := middleware.AuthenticatedUser{}.CurrentUser(r.Context())
	if signedIn {
		_, _ = io.WriteString(w, "Hello, "+user.Name)
	}
}

func TestSignInStoresTheUserWithTheFingerprintOfItsPasswordHash(t *testing.T) {
	t.Parallel()

	app := newAuthApp()

	assert.Equal(t, userID+" fingerprint of hash-1", app.serve(t, showingTheUser, app.signIn(t)).body)
}

func TestASignedInUserPassesTheGuardAsTheAuthenticatedUser(t *testing.T) {
	t.Parallel()

	app := newAuthApp()

	assert.Equal(t, "Hello, Ada Lovelace", app.requestTheGuardedPage(t, app.signIn(t)).body)
}

func TestAGuestIsSentToLogin(t *testing.T) {
	t.Parallel()

	app := newAuthApp()

	guarded := app.requestTheGuardedPage(t, nil)

	assert.Equal(t, http.StatusSeeOther, guarded.status)
	assert.Equal(t, []string{route.Login}, app.responder.redirects)
	assert.Empty(t, guarded.body)
}

func TestASessionWhoseUserIsGoneIsSignedOutAndSentToLogin(t *testing.T) {
	t.Parallel()

	app := newAuthApp()
	cookie := app.signIn(t)
	delete(app.users.users, userID)

	app.assertSignedOutAndSentToLogin(t, cookie)
}

func TestASessionFromBeforeAPasswordChangeIsSignedOutAndSentToLogin(t *testing.T) {
	t.Parallel()

	app := newAuthApp()
	cookie := app.signIn(t)
	changed := ada()
	changed.PasswordHash = "hash-2"
	app.users.users[userID] = changed

	app.assertSignedOutAndSentToLogin(t, cookie)
}

func TestAFailingUserLookupRendersTheError(t *testing.T) {
	t.Parallel()

	app := newAuthApp()
	cookie := app.signIn(t)
	app.users.err = errUsersDown

	guarded := app.requestTheGuardedPage(t, cookie)

	assert.Equal(t, http.StatusTeapot, guarded.status)
	assert.Len(t, app.responder.rendered, 1)
	assert.ErrorIs(t, app.responder.rendered[0], errUsersDown)
}

func (a *authApp) intendedAfterLogin(t *testing.T, cookie *http.Cookie) string {
	t.Helper()

	return a.serve(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, a.authenticator.Intended(r.Context()))
	}, cookie).body
}

func (a *authApp) guestRequests(t *testing.T, target string) *http.Cookie {
	t.Helper()

	return a.serveAt(t, target, a.authenticator.Require(http.HandlerFunc(greetingTheUser)).ServeHTTP, nil).cookie(t)
}

func TestTheGuardedPageAGuestAskedForIsIntendedOnceAfterLogin(t *testing.T) {
	t.Parallel()

	app := newAuthApp()
	cookie := app.guestRequests(t, "/dashboard?tab=team")

	assert.Equal(t, "/dashboard?tab=team", app.intendedAfterLogin(t, cookie))
	assert.Equal(t, route.Dashboard, app.intendedAfterLogin(t, cookie))
}

func TestWithoutAGuardedPageAskedForTheDashboardIsIntended(t *testing.T) {
	t.Parallel()

	assert.Equal(t, route.Dashboard, newAuthApp().intendedAfterLogin(t, nil))
}

func TestAnIntendedPageOffThisSiteFallsBackToTheDashboard(t *testing.T) {
	t.Parallel()

	app := newAuthApp()
	cookie := app.guestRequests(t, "//evil.example/")

	assert.Equal(t, route.Dashboard, app.intendedAfterLogin(t, cookie))
}

func (a *authApp) requestTheGuestPage(t *testing.T, cookie *http.Cookie) served {
	t.Helper()

	return a.serve(t, a.authenticator.RequireGuest(http.HandlerFunc(welcomingTheGuest)).ServeHTTP, cookie)
}

func welcomingTheGuest(w http.ResponseWriter, _ *http.Request) {
	_, _ = io.WriteString(w, "Welcome")
}

func TestAGuestPassesTheGuestGuard(t *testing.T) {
	t.Parallel()

	app := newAuthApp()

	assert.Equal(t, "Welcome", app.requestTheGuestPage(t, nil).body)
}

func TestASignedInUserIsSentFromAGuestPageToTheDashboard(t *testing.T) {
	t.Parallel()

	app := newAuthApp()

	guarded := app.requestTheGuestPage(t, app.signIn(t))

	assert.Equal(t, http.StatusSeeOther, guarded.status)
	assert.Equal(t, []string{route.Dashboard}, app.responder.redirects)
	assert.Empty(t, guarded.body)
}

func TestASessionFromBeforeAPasswordChangePassesTheGuestGuardSignedOut(t *testing.T) {
	t.Parallel()

	app := newAuthApp()
	cookie := app.signIn(t)
	changed := ada()
	changed.PasswordHash = "hash-2"
	app.users.users[userID] = changed

	assert.Equal(t, "Welcome", app.requestTheGuestPage(t, cookie).body)
	assert.Equal(t, "guest ", app.serve(t, showingTheUser, cookie).body)
}

func TestAFailingUserLookupOnAGuestPageRendersTheError(t *testing.T) {
	t.Parallel()

	app := newAuthApp()
	cookie := app.signIn(t)
	app.users.err = errUsersDown

	assert.Equal(t, http.StatusTeapot, app.requestTheGuestPage(t, cookie).status)
	assert.Len(t, app.responder.rendered, 1)
}

func (a *authApp) assertSignedOutAndSentToLogin(t *testing.T, cookie *http.Cookie) {
	t.Helper()

	guarded := a.requestTheGuardedPage(t, cookie)

	assert.Equal(t, http.StatusSeeOther, guarded.status)
	assert.Equal(t, []string{route.Login}, a.responder.redirects)
	assert.NotEqual(t, cookie.Value, guarded.cookie(t).Value, "the session token was not replaced")
	assert.Equal(t, "guest ", a.serve(t, showingTheUser, cookie).body)
}
