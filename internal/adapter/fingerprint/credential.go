package fingerprint

import (
	"crypto/hmac"
	"encoding/hex"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type digester interface {
	Digest(message []byte) []byte
}

type Credentials struct {
	digester digester
}

func NewCredentials(digester digester) *Credentials {
	return &Credentials{digester: digester}
}

func (c *Credentials) Of(hash domain.PasswordHash) string {
	return hex.EncodeToString(c.digester.Digest([]byte(hash)))
}

func (c *Credentials) Matches(hash domain.PasswordHash, fingerprint string) bool {
	return hmac.Equal([]byte(c.Of(hash)), []byte(fingerprint))
}
