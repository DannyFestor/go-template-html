package verificationlink

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"time"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type temporarySigner interface {
	SignTemporaryFor(path string, query url.Values, ttl time.Duration) string
}

type Links struct {
	signer temporarySigner
	path   string
	ttl    time.Duration
}

func NewLinks(signer temporarySigner, path string, ttl time.Duration) *Links {
	return &Links{signer: signer, path: path, ttl: ttl}
}

func (l *Links) For(user domain.User) string {
	return l.signer.SignTemporaryFor(l.pathFor(user), url.Values{}, l.ttl)
}

func (l *Links) Matches(email, hash string) bool {
	return hmac.Equal([]byte(emailHash(email)), []byte(hash))
}

func (l *Links) pathFor(user domain.User) string {
	return l.path + "/" + url.PathEscape(string(user.ID)) + "/" + emailHash(user.Email)
}

func emailHash(email string) string {
	sum := sha256.Sum256([]byte(email))

	return hex.EncodeToString(sum[:])
}
