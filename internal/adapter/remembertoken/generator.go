package remembertoken

import (
	"crypto/rand"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type Generator struct{}

func (Generator) New() domain.RememberToken {
	return domain.RememberToken{Selector: rand.Text(), Validator: rand.Text()}
}
