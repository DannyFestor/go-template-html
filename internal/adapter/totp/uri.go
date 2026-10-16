package totp

import (
	"encoding/base32"
	"net/url"
	"strings"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

func URI(issuer, account string, secret domain.TwoFactorSecret) string {
	return "otpauth://totp/" + url.PathEscape(issuer+":"+account) +
		"?secret=" + base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret) +
		"&issuer=" + escapeForAppsReadingPlusLiterally(issuer)
}

func escapeForAppsReadingPlusLiterally(value string) string {
	return strings.ReplaceAll(url.QueryEscape(value), "+", "%20")
}
