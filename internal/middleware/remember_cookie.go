package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

const (
	rememberCookieName     = "remember"
	rememberTokenSeparator = "."
)

type RememberCookie struct {
	cookie hostOnlyCookie
	maxAge int
}

func NewRememberCookie(secure bool, lifetime time.Duration) RememberCookie {
	return RememberCookie{
		cookie: hostOnlyCookie{baseName: rememberCookieName, secure: secure},
		maxAge: int(lifetime.Seconds()),
	}
}

func (c RememberCookie) Read(r *http.Request) (domain.RememberToken, bool) {
	var none domain.RememberToken

	value, found := c.cookie.read(r)
	if !found {
		return none, false
	}

	selector, validator, separated := strings.Cut(value, rememberTokenSeparator)
	if !separated || selector == "" || validator == "" {
		return none, false
	}

	return domain.RememberToken{Selector: selector, Validator: validator}, true
}

func (c RememberCookie) Issue(w http.ResponseWriter, token domain.RememberToken) {
	c.cookie.set(w, token.Selector+rememberTokenSeparator+token.Validator, c.maxAge)
}

func (c RememberCookie) Clear(w http.ResponseWriter) {
	c.cookie.clear(w)
}
