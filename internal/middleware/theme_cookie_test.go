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

const fourHundredDaysInSeconds = 400 * 24 * 60 * 60

func issueThemeCookie(t *testing.T, secure bool, theme domain.Theme) *http.Cookie {
	t.Helper()

	recorder := httptest.NewRecorder()
	middleware.NewThemeCookie(secure).Issue(recorder, theme)

	cookies := recorder.Result().Cookies()
	require.Len(t, cookies, 1)

	return cookies[0]
}

func readThemeCookie(t *testing.T, secure bool, header string) domain.Theme {
	t.Helper()

	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	if header != "" {
		request.Header.Set("Cookie", header)
	}

	return middleware.NewThemeCookie(secure).Read(request)
}

func TestTheThemeCookieLastsAsLongAsBrowsersKeepCookiesOnThisHost(t *testing.T) {
	t.Parallel()

	cookie := issueThemeCookie(t, false, domain.ThemeDark)

	assert.Equal(t, "theme", cookie.Name)
	assert.Equal(t, "dark", cookie.Value)
	assert.Equal(t, fourHundredDaysInSeconds, cookie.MaxAge)
	assert.Equal(t, "/", cookie.Path)
	assert.Empty(t, cookie.Domain)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	assert.False(t, cookie.Secure)
}

func TestTheThemeCookieIsHostPrefixedAndSecureOverHTTPS(t *testing.T) {
	t.Parallel()

	cookie := issueThemeCookie(t, true, domain.ThemeLight)

	assert.Equal(t, "__Host-theme", cookie.Name)
	assert.True(t, cookie.Secure)
}

func TestTheThemeCookieReadsBackTheThemeItCarries(t *testing.T) {
	t.Parallel()

	for _, secure := range []bool{false, true} {
		for _, theme := range []domain.Theme{domain.ThemeLight, domain.ThemeDark, domain.ThemeSystem} {
			cookie := issueThemeCookie(t, secure, theme)

			assert.Equal(t, theme, readThemeCookie(t, secure, cookie.Name+"="+cookie.Value), theme)
		}
	}
}

func TestWithoutAKnownThemeCookieTheThemeFollowsTheSystem(t *testing.T) {
	t.Parallel()

	for _, header := range []string{"", "theme=", "theme=sepia", "__Host-theme=dark"} {
		assert.Equal(t, domain.ThemeSystem, readThemeCookie(t, false, header), header)
	}
}
