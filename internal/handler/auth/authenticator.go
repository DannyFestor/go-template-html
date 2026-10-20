package auth

import (
	"context"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type authenticator interface {
	SignIn(ctx context.Context, user domain.User)
	SignOut(ctx context.Context)
	SignedIn(ctx context.Context) bool
	Intended(ctx context.Context) string
}
