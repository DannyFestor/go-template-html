package totp

import (
	"crypto/subtle"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

const driftSteps = 1

type Authenticator struct {
	clock clock
}

func NewAuthenticator(clock clock) *Authenticator {
	return &Authenticator{clock: clock}
}

func (a *Authenticator) Match(secret domain.TwoFactorSecret, code string) (domain.TimeStep, bool) {
	current := stepAt(a.clock.Now())

	for step := current + driftSteps; step >= current-driftSteps; step-- {
		if subtle.ConstantTimeCompare([]byte(codeAt(secret, step)), []byte(code)) == 1 {
			return step, true
		}
	}

	return 0, false
}
