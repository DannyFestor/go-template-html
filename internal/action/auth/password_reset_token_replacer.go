package auth

import (
	"context"
	"time"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

//go:generate go tool mockgen -source=password_reset_token_replacer.go -destination=mock_password_reset_token_replacer_test.go -package=auth -mock_names=passwordResetTokenReplacer=MockPasswordResetTokenReplacer

type passwordResetTokenReplacer interface {
	Replace(ctx context.Context, userID domain.UserID, token domain.PasswordResetToken, expiresAt time.Time) error
}
