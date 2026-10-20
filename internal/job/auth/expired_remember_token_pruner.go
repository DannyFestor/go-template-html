package auth

import (
	"context"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
)

//go:generate go tool mockgen -source=expired_remember_token_pruner.go -destination=mock_expired_remember_token_pruner_test.go -package=auth -mock_names=expiredRememberTokenPruner=MockExpiredRememberTokenPruner

type expiredRememberTokenPruner interface {
	Run(ctx context.Context, in auth.PruneExpiredRememberTokensIn) (auth.PruneExpiredRememberTokensOut, error)
}
