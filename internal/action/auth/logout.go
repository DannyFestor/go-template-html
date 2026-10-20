package auth

import (
	"context"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type Logout struct {
	tokens rememberTokenForgetter
}

type LogoutIn struct {
	RememberToken domain.RememberToken
}

type LogoutOut struct{}

func NewLogout(tokens rememberTokenForgetter) *Logout {
	return &Logout{tokens: tokens}
}

func (l *Logout) Run(ctx context.Context, in LogoutIn) (LogoutOut, error) {
	err := l.tokens.Forget(ctx, in.RememberToken)
	if err != nil {
		return LogoutOut{}, fmt.Errorf("log out: %w", err)
	}

	return LogoutOut{}, nil
}
