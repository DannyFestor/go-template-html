package auth

import (
	"context"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
)

type PruneExpiredRememberTokens struct{}

func (PruneExpiredRememberTokens) Kind() string {
	return "prune_expired_remember_tokens"
}

type PruneExpiredRememberTokensHandler struct {
	prune expiredRememberTokenPruner
}

func NewPruneExpiredRememberTokensHandler(prune expiredRememberTokenPruner) *PruneExpiredRememberTokensHandler {
	return &PruneExpiredRememberTokensHandler{prune: prune}
}

func (h *PruneExpiredRememberTokensHandler) Handle(ctx context.Context, _ PruneExpiredRememberTokens) error {
	_, err := h.prune.Run(ctx, auth.PruneExpiredRememberTokensIn{})
	if err != nil {
		return fmt.Errorf("handle the %s job: %w", PruneExpiredRememberTokens{}.Kind(), err)
	}

	return nil
}
