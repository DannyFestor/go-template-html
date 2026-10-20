package auth

import (
	"context"
	"time"
)

//go:generate go tool mockgen -source=expired_remember_token_deleter.go -destination=mock_expired_remember_token_deleter_test.go -package=auth -mock_names=expiredRememberTokenDeleter=MockExpiredRememberTokenDeleter

type expiredRememberTokenDeleter interface {
	DeleteExpired(ctx context.Context, now time.Time) error
}
