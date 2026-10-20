package middleware_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/middleware"
)

const (
	absoluteLifetime = 24 * time.Hour
	userID           = "019a1f6e-3c4b-7d2e-8f10-1234567890ab"
	fingerprint      = "fp-1"
	flash            = "Profile saved."
)

var errStoreDown = errors.New("store down")

type memoryStore struct {
	mutex    sync.Mutex
	sessions map[string][]byte
	saves    int
	loadErr  error
	saveErr  error
}

func newMemoryStore() *memoryStore {
	return &memoryStore{mutex: sync.Mutex{}, sessions: map[string][]byte{}, saves: 0, loadErr: nil, saveErr: nil}
}

func (s *memoryStore) Load(_ context.Context, id string) ([]byte, bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	data, found := s.sessions[id]

	return data, found, s.loadErr
}

func (s *memoryStore) Save(_ context.Context, id string, data []byte) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.saveErr != nil {
		return s.saveErr
	}

	s.saves++
	s.sessions[id] = data

	return nil
}

func (s *memoryStore) Delete(_ context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.sessions, id)

	return nil
}

func (s *memoryStore) saveCount() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.saves
}

type manualClock struct {
	now time.Time
}

func (c *manualClock) Now() time.Time {
	return c.now
}

type sessionApp struct {
	store    *memoryStore
	clock    *manualClock
	secure   bool
	renderer *spyRenderer
}

type served struct {
	status  int
	cookies []*http.Cookie
	body    string
}

func newSessionApp() *sessionApp {
	return &sessionApp{
		store:    newMemoryStore(),
		clock:    &manualClock{now: time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)},
		secure:   false,
		renderer: &spyRenderer{rendered: nil},
	}
}

func (a *sessionApp) serve(t *testing.T, handler http.HandlerFunc, cookie *http.Cookie) served {
	t.Helper()

	return a.serveAt(t, "/", handler, cookie)
}

func (a *sessionApp) serveAt(t *testing.T, target string, handler http.HandlerFunc, cookie *http.Cookie) served {
	t.Helper()

	settings := middleware.SessionSettings{Secure: a.secure, AbsoluteLifetime: absoluteLifetime}

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
	if cookie != nil {
		request.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()
	middleware.LoadAndSaveSession(a.store, a.clock, settings, a.renderer)(handler).ServeHTTP(recorder, request)

	return served{status: recorder.Code, cookies: recorder.Result().Cookies(), body: recorder.Body.String()}
}

func (s served) cookie(t *testing.T) *http.Cookie {
	t.Helper()

	require.Len(t, s.cookies, 1)

	return s.cookies[0]
}

func signingIn(w http.ResponseWriter, r *http.Request) {
	middleware.Session{}.SignIn(r.Context(), userID, fingerprint)
	w.WriteHeader(http.StatusNoContent)
}

func signingOut(w http.ResponseWriter, r *http.Request) {
	middleware.Session{}.SignOut(r.Context())
	w.WriteHeader(http.StatusNoContent)
}

func flashing(w http.ResponseWriter, r *http.Request) {
	middleware.Session{}.SetFlash(r.Context(), flash)
	w.WriteHeader(http.StatusNoContent)
}

func takingTheFlash(w http.ResponseWriter, r *http.Request) {
	_, _ = io.WriteString(w, middleware.Session{}.TakeFlash(r.Context()))
}

func showingTheUser(w http.ResponseWriter, r *http.Request) {
	current, signedIn := middleware.Session{}.CurrentUserID(r.Context())
	if !signedIn {
		current = "guest"
	}

	_, _ = io.WriteString(w, current+" "+middleware.Session{}.CredentialFingerprint(r.Context()))
}

func TestASignedInUserIsKnownOnTheNextRequest(t *testing.T) {
	t.Parallel()

	app := newSessionApp()
	cookie := app.serve(t, signingIn, nil).cookie(t)

	assert.Equal(t, userID+" "+fingerprint, app.serve(t, showingTheUser, cookie).body)
}

func TestARequestThatLeavesTheSessionAloneGetsNoCookie(t *testing.T) {
	t.Parallel()

	app := newSessionApp()

	response := app.serve(t, showingTheUser, nil)

	assert.Equal(t, "guest ", response.body)
	assert.Empty(t, response.cookies)
	assert.Zero(t, app.store.saves)
}

func TestTheSessionCookieStaysOnThisHostAndAwayFromScripts(t *testing.T) {
	t.Parallel()

	cookie := newSessionApp().serve(t, signingIn, nil).cookie(t)

	assert.Equal(t, "session", cookie.Name)
	assert.Equal(t, "/", cookie.Path)
	assert.Empty(t, cookie.Domain)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	assert.False(t, cookie.Secure)
	assert.Zero(t, cookie.MaxAge, "a persistent cookie")
}

func TestTheSessionCookieIsHostPrefixedAndSecureOverHTTPS(t *testing.T) {
	t.Parallel()

	app := newSessionApp()
	app.secure = true

	cookie := app.serve(t, signingIn, nil).cookie(t)

	assert.Equal(t, "__Host-session", cookie.Name)
	assert.True(t, cookie.Secure)
	assert.Equal(t, userID+" "+fingerprint, app.serve(t, showingTheUser, cookie).body)
}

func TestReadingASessionNeitherSavesItNorSendsTheCookieAgain(t *testing.T) {
	t.Parallel()

	app := newSessionApp()
	cookie := app.serve(t, signingIn, nil).cookie(t)

	response := app.serve(t, showingTheUser, cookie)

	assert.Equal(t, 1, app.store.saves)
	assert.Empty(t, response.cookies)
}

func TestSigningInAgainReplacesTheSessionToken(t *testing.T) {
	t.Parallel()

	app := newSessionApp()
	first := app.serve(t, signingIn, nil).cookie(t)

	second := app.serve(t, signingIn, first).cookie(t)

	assert.NotEqual(t, first.Value, second.Value)
	assert.Equal(t, "guest ", app.serve(t, showingTheUser, first).body)
	assert.Equal(t, userID+" "+fingerprint, app.serve(t, showingTheUser, second).body)
}

func TestSigningOutEndsTheSessionAndClearsTheCookie(t *testing.T) {
	t.Parallel()

	app := newSessionApp()
	cookie := app.serve(t, signingIn, nil).cookie(t)

	cleared := app.serve(t, signingOut, cookie).cookie(t)

	assert.Equal(t, "session", cleared.Name)
	assert.Negative(t, cleared.MaxAge)
	assert.Equal(t, "guest ", app.serve(t, showingTheUser, cookie).body)
}

func TestASessionEndsAtItsAbsoluteLifetime(t *testing.T) {
	t.Parallel()

	app := newSessionApp()
	cookie := app.serve(t, signingIn, nil).cookie(t)

	app.clock.now = app.clock.now.Add(absoluteLifetime - time.Nanosecond)
	assert.Equal(t, userID+" "+fingerprint, app.serve(t, showingTheUser, cookie).body)

	app.clock.now = app.clock.now.Add(time.Nanosecond)
	assert.Equal(t, "guest ", app.serve(t, showingTheUser, cookie).body)
}

func TestASessionPastItsAbsoluteLifetimeIsDeletedAndItsCookieCleared(t *testing.T) {
	t.Parallel()

	app := newSessionApp()
	cookie := app.serve(t, signingIn, nil).cookie(t)
	app.clock.now = app.clock.now.Add(absoluteLifetime)

	response := app.serve(t, showingTheUser, cookie)

	assert.Negative(t, response.cookie(t).MaxAge)
	assert.Empty(t, app.store.sessions)
}

func TestACookieWithoutAStoredSessionIsCleared(t *testing.T) {
	t.Parallel()

	app := newSessionApp()
	cookie := app.serve(t, signingIn, nil).cookie(t)
	app.store.sessions = map[string][]byte{}

	assert.Negative(t, app.serve(t, showingTheUser, cookie).cookie(t).MaxAge)
}

func TestAFlashIsShownOnceThenTheGuestSessionEnds(t *testing.T) {
	t.Parallel()

	app := newSessionApp()
	cookie := app.serve(t, flashing, nil).cookie(t)

	shown := app.serve(t, takingTheFlash, cookie)

	assert.Equal(t, flash, shown.body)
	assert.Negative(t, shown.cookie(t).MaxAge)
	assert.Empty(t, app.serve(t, takingTheFlash, cookie).body)
}

func TestTheCookieIsSetWhenTheHandlerWritesBeforeItsHeaders(t *testing.T) {
	t.Parallel()

	handlers := map[string]http.HandlerFunc{
		"write": func(w http.ResponseWriter, r *http.Request) {
			middleware.Session{}.SignIn(r.Context(), userID, fingerprint)

			_, _ = io.WriteString(w, "streamed")
		},
		"flush": func(w http.ResponseWriter, r *http.Request) {
			middleware.Session{}.SignIn(r.Context(), userID, fingerprint)

			_ = http.NewResponseController(w).Flush()
		},
	}

	for name, handler := range handlers {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			app := newSessionApp()
			cookie := app.serve(t, handler, nil).cookie(t)

			assert.Equal(t, userID+" "+fingerprint, app.serve(t, showingTheUser, cookie).body)
		})
	}
}

func TestTheSessionIsSavedBeforeTheHandlerHijacksTheConnection(t *testing.T) {
	t.Parallel()

	app := newSessionApp()
	settings := middleware.SessionSettings{Secure: false, AbsoluteLifetime: absoluteLifetime}
	hijacked := make(chan int, 1)
	server := httptest.NewServer(middleware.LoadAndSaveSession(app.store, app.clock, settings, app.renderer)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middleware.Session{}.SignIn(r.Context(), userID, fingerprint)

			conn, _, err := http.NewResponseController(w).Hijack()
			if err == nil {
				_ = conn.Close()
			}

			hijacked <- app.store.saveCount()
		}),
	))
	t.Cleanup(server.Close)

	response, err := server.Client().Do(newServerRequest(t, server.URL))
	if err == nil {
		_ = response.Body.Close()
	}

	assert.Equal(t, 1, <-hijacked)
}

func TestAFailingLoadRendersTheErrorWithoutRunningTheHandler(t *testing.T) {
	t.Parallel()

	app := newSessionApp()
	cookie := app.serve(t, signingIn, nil).cookie(t)
	app.store.loadErr = errStoreDown

	response := app.serve(t, showingTheUser, cookie)

	assert.Equal(t, http.StatusTeapot, response.status)
	assert.Empty(t, response.body)
	require.Len(t, app.renderer.rendered, 1)
	assert.ErrorIs(t, app.renderer.rendered[0], errStoreDown)
}

func TestAFailingSaveRendersTheErrorInsteadOfTheResponse(t *testing.T) {
	t.Parallel()

	app := newSessionApp()
	app.store.saveErr = errStoreDown

	response := app.serve(t, func(w http.ResponseWriter, r *http.Request) {
		middleware.Session{}.SignIn(r.Context(), userID, fingerprint)

		_, _ = io.WriteString(w, "signed in")
	}, nil)

	assert.Equal(t, http.StatusTeapot, response.status)
	assert.Empty(t, response.body)
	assert.Empty(t, response.cookies)
	require.Len(t, app.renderer.rendered, 1)
	assert.ErrorIs(t, app.renderer.rendered[0], errStoreDown)
}

func TestTheSessionReadsAsEmptyOutsideTheMiddleware(t *testing.T) {
	t.Parallel()

	current, signedIn := middleware.Session{}.CurrentUserID(t.Context())

	assert.Empty(t, current)
	assert.False(t, signedIn)
	assert.Empty(t, middleware.Session{}.TakeFlash(t.Context()))
	assert.Panics(t, func() { middleware.Session{}.SignIn(t.Context(), userID, fingerprint) })
}
