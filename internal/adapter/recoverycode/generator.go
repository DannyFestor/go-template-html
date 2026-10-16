package recoverycode

import (
	"crypto/rand"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

const (
	setSize    = 8
	groupChars = 10
)

func NewSet() []domain.RecoveryCode {
	codes := make([]domain.RecoveryCode, 0, setSize)
	for range setSize {
		codes = append(codes, newCode())
	}

	return codes
}

func newCode() domain.RecoveryCode {
	random := rand.Text()

	return domain.RecoveryCode(random[:groupChars] + "-" + random[groupChars:2*groupChars])
}
