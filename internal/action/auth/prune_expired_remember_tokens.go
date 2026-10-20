package auth

import (
	"context"
	"fmt"
)

type PruneExpiredRememberTokens struct {
	tokens expiredRememberTokenDeleter
	clock  clock
}

type PruneExpiredRememberTokensIn struct{}

type PruneExpiredRememberTokensOut struct{}

func NewPruneExpiredRememberTokens(tokens expiredRememberTokenDeleter, clock clock) *PruneExpiredRememberTokens {
	return &PruneExpiredRememberTokens{tokens: tokens, clock: clock}
}

func (p *PruneExpiredRememberTokens) Run(
	ctx context.Context,
	_ PruneExpiredRememberTokensIn,
) (PruneExpiredRememberTokensOut, error) {
	err := p.tokens.DeleteExpired(ctx, p.clock.Now())
	if err != nil {
		return PruneExpiredRememberTokensOut{}, fmt.Errorf("prune the expired remember tokens: %w", err)
	}

	return PruneExpiredRememberTokensOut{}, nil
}
