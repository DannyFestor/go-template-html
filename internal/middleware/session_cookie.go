package middleware

import (
	"net/http"
	"time"
)

const (
	sessionCookieName = "session"
	hostOnlyPrefix    = "__Host-"
)

type sessionCookie struct {
	secure bool
}

func (c sessionCookie) name() string {
	if c.secure {
		return hostOnlyPrefix + sessionCookieName
	}

	return sessionCookieName
}

func (c sessionCookie) read(r *http.Request) (sessionToken, bool) {
	cookie, err := r.Cookie(c.name())
	if err != nil || cookie.Value == "" {
		return "", false
	}

	return sessionToken(cookie.Value), true
}

func (c sessionCookie) issue(w http.ResponseWriter, token sessionToken) {
	http.SetCookie(w, c.cookie(string(token), 0))
}

func (c sessionCookie) clear(w http.ResponseWriter) {
	http.SetCookie(w, c.cookie("", -1))
}

func (c sessionCookie) cookie(value string, maxAge int) *http.Cookie {
	//nolint:gosec // Secure follows APP_URL, because browsers drop Secure cookies set over plain http off localhost.
	return &http.Cookie{
		Name:        c.name(),
		Value:       value,
		Quoted:      false,
		Path:        "/",
		Domain:      "",
		Expires:     time.Time{},
		RawExpires:  "",
		MaxAge:      maxAge,
		Secure:      c.secure,
		HttpOnly:    true,
		SameSite:    http.SameSiteLaxMode,
		Partitioned: false,
		Raw:         "",
		Unparsed:    nil,
	}
}
