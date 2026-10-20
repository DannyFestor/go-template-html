package auth

import (
	"context"
	"time"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

//go:generate go tool mockgen -source=remember_token_consumer.go -destination=mock_remember_token_consumer_test.go -package=auth -mock_names=rememberTokenConsumer=MockRememberTokenConsumer

type rememberTokenConsumer interface {
	Consume(ctx context.Context, token domain.RememberToken, now time.Time) (domain.RememberTokenUse, bool, error)
}
