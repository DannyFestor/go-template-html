package totp

import (
	"crypto/rand"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

const rfc4226RecommendedSecretBytes = 20

func NewSecret() domain.TwoFactorSecret {
	secret := make(domain.TwoFactorSecret, rfc4226RecommendedSecretBytes)
	_, _ = rand.Read(secret)

	return secret
}
