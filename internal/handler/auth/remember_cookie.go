package auth

import (
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type rememberCookie interface {
	Read(r *http.Request) (domain.RememberToken, bool)
	Issue(w http.ResponseWriter, token domain.RememberToken)
	Clear(w http.ResponseWriter)
}
