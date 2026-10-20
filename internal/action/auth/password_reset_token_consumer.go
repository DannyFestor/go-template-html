package auth

import (
	"context"
	"time"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

//go:generate go tool mockgen -source=password_reset_token_consumer.go -destination=mock_password_reset_token_consumer_test.go -package=auth -mock_names=passwordResetTokenConsumer=MockPasswordResetTokenConsumer

type passwordResetTokenConsumer interface {
	Consume(ctx context.Context, token domain.PasswordResetToken, now time.Time) (domain.UserID, bool, error)
}
