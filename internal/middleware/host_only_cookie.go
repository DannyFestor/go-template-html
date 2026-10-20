package middleware

import (
	"net/http"
	"time"
)

const hostOnlyPrefix = "__Host-"

type hostOnlyCookie struct {
	baseName string
	secure   bool
}

func (c hostOnlyCookie) name() string {
	if c.secure {
		return hostOnlyPrefix + c.baseName
	}

	return c.baseName
}

func (c hostOnlyCookie) read(r *http.Request) (string, bool) {
	cookie, err := r.Cookie(c.name())
	if err != nil || cookie.Value == "" {
		return "", false
	}

	return cookie.Value, true
}

func (c hostOnlyCookie) set(w http.ResponseWriter, value string, maxAge int) {
	http.SetCookie(w, c.cookie(value, maxAge))
}

func (c hostOnlyCookie) clear(w http.ResponseWriter) {
	http.SetCookie(w, c.cookie("", -1))
}

func (c hostOnlyCookie) cookie(value string, maxAge int) *http.Cookie {
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
