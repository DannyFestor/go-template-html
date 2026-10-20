package auth

import (
	"context"
	"time"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

//go:generate go tool mockgen -source=remember_token_issuer.go -destination=mock_remember_token_issuer_test.go -package=auth -mock_names=rememberTokenIssuer=MockRememberTokenIssuer

type rememberTokenIssuer interface {
	Issue(ctx context.Context, userID domain.UserID, token domain.RememberToken, expiresAt time.Time) error
}
