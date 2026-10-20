package middleware

import (
	"net/http"
	"time"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

const (
	themeCookieName           = "theme"
	browsersLongestCookieLife = 400 * 24 * time.Hour
)

type ThemeCookie struct {
	cookie hostOnlyCookie
}

func NewThemeCookie(secure bool) ThemeCookie {
	return ThemeCookie{cookie: hostOnlyCookie{baseName: themeCookieName, secure: secure}}
}

func (c ThemeCookie) Read(r *http.Request) domain.Theme {
	value, _ := c.cookie.read(r)

	theme := domain.Theme(value)
	if !theme.Valid() {
		return domain.ThemeSystem
	}

	return theme
}

func (c ThemeCookie) Issue(w http.ResponseWriter, theme domain.Theme) {
	c.cookie.set(w, string(theme), int(browsersLongestCookieLife.Seconds()))
}
