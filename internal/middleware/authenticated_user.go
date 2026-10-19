package middleware

import (
	"context"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type authenticatedUserKey struct{}

type AuthenticatedUser struct{}

func (AuthenticatedUser) CurrentUser(ctx context.Context) (domain.User, bool) {
	user, ok := ctx.Value(authenticatedUserKey{}).(domain.User)

	return user, ok
}

func withAuthenticatedUser(ctx context.Context, user domain.User) context.Context {
	return context.WithValue(ctx, authenticatedUserKey{}, user)
}
