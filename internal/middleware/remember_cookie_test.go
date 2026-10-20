package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/middleware"
)

const rememberLifetime = 30 * 24 * time.Hour

func rememberToken() domain.RememberToken {
	return domain.RememberToken{Selector: "SELECTOR", Validator: "VALIDATOR"}
}

func issueRememberCookie(t *testing.T, secure bool) *http.Cookie {
	t.Helper()

	recorder := httptest.NewRecorder()
	middleware.NewRememberCookie(secure, rememberLifetime).Issue(recorder, rememberToken())

	cookies := recorder.Result().Cookies()
	require.Len(t, cookies, 1)

	return cookies[0]
}

func readRememberCookie(t *testing.T, secure bool, cookie *http.Cookie) (domain.RememberToken, bool) {
	t.Helper()

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	request.AddCookie(cookie)

	return middleware.NewRememberCookie(secure, rememberLifetime).Read(request)
}

func TestTheRememberCookieLastsTheRememberLifetimeOnThisHostAwayFromScripts(t *testing.T) {
	t.Parallel()

	cookie := issueRememberCookie(t, false)

	assert.Equal(t, "remember", cookie.Name)
	assert.Equal(t, int(rememberLifetime.Seconds()), cookie.MaxAge)
	assert.Equal(t, "/", cookie.Path)
	assert.Empty(t, cookie.Domain)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	assert.False(t, cookie.Secure)
}

func TestTheRememberCookieIsHostPrefixedAndSecureOverHTTPS(t *testing.T) {
	t.Parallel()

	cookie := issueRememberCookie(t, true)

	assert.Equal(t, "__Host-remember", cookie.Name)
	assert.True(t, cookie.Secure)
}

func TestTheRememberCookieReadsBackTheTokenItCarries(t *testing.T) {
	t.Parallel()

	for _, secure := range []bool{false, true} {
		token, found := readRememberCookie(t, secure, issueRememberCookie(t, secure))

		assert.True(t, found)
		assert.Equal(t, rememberToken(), token)
	}
}

func TestAMalformedRememberCookieCarriesNoToken(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"", "SELECTOR", ".VALIDATOR", "SELECTOR."} {
		request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
		request.Header.Set("Cookie", "remember="+value)

		_, found := middleware.NewRememberCookie(false, rememberLifetime).Read(request)

		assert.False(t, found, value)
	}
}

func TestClearingTheRememberCookieExpiresIt(t *testing.T) {
	t.Parallel()

	recorder := httptest.NewRecorder()
	middleware.NewRememberCookie(false, rememberLifetime).Clear(recorder)

	cookies := recorder.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "remember", cookies[0].Name)
	assert.Negative(t, cookies[0].MaxAge)
}
