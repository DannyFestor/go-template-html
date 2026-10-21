package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/middleware"
)

func (a *authApp) readSignedInUser(t *testing.T, cookie *http.Cookie) (domain.User, bool, error) {
	t.Helper()

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/ws", nil)
	if cookie != nil {
		request.AddCookie(cookie)
	}

	settings := middleware.SessionSettings{Secure: a.secure, AbsoluteLifetime: absoluteLifetime}

	//nolint:wrapcheck // Tests assert on the error itself.
	return middleware.NewSessionReader(a.store, a.clock, settings, a.authenticator).SignedInUser(request)
}

func TestTheSessionReaderFindsTheSignedInUser(t *testing.T) {
	t.Parallel()

	app := newAuthApp()

	user, signedIn, err := app.readSignedInUser(t, app.signIn(t))

	require.NoError(t, err)
	assert.True(t, signedIn)
	assert.Equal(t, ada(), user)
}

func TestTheSessionReaderSeesAGuestWithoutACookie(t *testing.T) {
	t.Parallel()

	_, signedIn, err := newAuthApp().readSignedInUser(t, nil)

	require.NoError(t, err)
	assert.False(t, signedIn)
}

func TestTheSessionReaderSeesAGuestOnceTheSessionIsGone(t *testing.T) {
	t.Parallel()

	app := newAuthApp()
	cookie := app.signIn(t)
	app.serve(t, signingOut, cookie)

	_, signedIn, err := app.readSignedInUser(t, cookie)

	require.NoError(t, err)
	assert.False(t, signedIn)
}

func TestTheSessionReaderSeesAGuestOnceThePasswordChanged(t *testing.T) {
	t.Parallel()

	app := newAuthApp()
	cookie := app.signIn(t)
	changed := ada()
	changed.PasswordHash = "hash-2"
	app.users.users[userID] = changed

	_, signedIn, err := app.readSignedInUser(t, cookie)

	require.NoError(t, err)
	assert.False(t, signedIn)
}

func TestTheSessionReaderSeesAGuestPastTheAbsoluteLifetime(t *testing.T) {
	t.Parallel()

	app := newAuthApp()
	cookie := app.signIn(t)
	app.clock.now = app.clock.now.Add(absoluteLifetime)

	_, signedIn, err := app.readSignedInUser(t, cookie)

	require.NoError(t, err)
	assert.False(t, signedIn)
}

func TestTheSessionReaderLeavesTheSessionAsItWas(t *testing.T) {
	t.Parallel()

	app := newAuthApp()
	cookie := app.signIn(t)
	changed := ada()
	changed.PasswordHash = "hash-2"
	app.users.users[userID] = changed
	saves := app.store.saveCount()

	_, _, err := app.readSignedInUser(t, cookie)

	require.NoError(t, err)
	assert.Equal(t, saves, app.store.saveCount())
	assert.Equal(t, userID+" fingerprint of hash-1", app.serve(t, showingTheUser, cookie).body)
}

func TestTheSessionReaderFailsWhenTheStoreOrTheUsersFail(t *testing.T) {
	t.Parallel()

	storeDown := newAuthApp()
	storeCookie := storeDown.signIn(t)
	storeDown.store.loadErr = errStoreDown

	usersDown := newAuthApp()
	usersCookie := usersDown.signIn(t)
	usersDown.users.err = errUsersDown

	_, _, err := storeDown.readSignedInUser(t, storeCookie)
	require.ErrorIs(t, err, errStoreDown)

	_, _, err = usersDown.readSignedInUser(t, usersCookie)
	assert.ErrorIs(t, err, errUsersDown)
}
