package resettoken

import (
	"crypto/rand"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type Generator struct{}

func (Generator) New() domain.PasswordResetToken {
	return domain.PasswordResetToken(rand.Text())
}
