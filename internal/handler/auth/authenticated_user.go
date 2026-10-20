package auth

import (
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/middleware"
)

func authenticatedUser(r *http.Request) domain.User {
	user, _ := middleware.AuthenticatedUser{}.CurrentUser(r.Context())

	return user
}
