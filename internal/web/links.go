package web

import (
	"net/url"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/route"
)

type Links struct {
	base url.URL
}

func NewLinks(base url.URL) *Links {
	return &Links{base: base}
}

func (l *Links) PasswordResetURL(token domain.PasswordResetToken) string {
	link := l.base.JoinPath(route.ResetPassword)
	link.RawQuery = url.Values{route.ResetPasswordTokenParam: {string(token)}}.Encode()

	return link.String()
}
