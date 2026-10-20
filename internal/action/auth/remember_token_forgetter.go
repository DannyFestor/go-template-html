package auth

import (
	"context"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

//go:generate go tool mockgen -source=remember_token_forgetter.go -destination=mock_remember_token_forgetter_test.go -package=auth -mock_names=rememberTokenForgetter=MockRememberTokenForgetter

type rememberTokenForgetter interface {
	Forget(ctx context.Context, token domain.RememberToken) error
}
