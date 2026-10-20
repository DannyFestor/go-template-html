package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type RememberDevice struct {
	generator rememberTokenGenerator
	tokens    rememberTokenIssuer
	clock     clock
	lifetime  time.Duration
}

type RememberDeviceIn struct {
	UserID domain.UserID
}

type RememberDeviceOut struct {
	Token domain.RememberToken
}

func NewRememberDevice(
	generator rememberTokenGenerator,
	tokens rememberTokenIssuer,
	clock clock,
	lifetime time.Duration,
) *RememberDevice {
	return &RememberDevice{generator: generator, tokens: tokens, clock: clock, lifetime: lifetime}
}

func (r *RememberDevice) Run(ctx context.Context, in RememberDeviceIn) (RememberDeviceOut, error) {
	token := r.generator.New()

	err := r.tokens.Issue(ctx, in.UserID, token, r.clock.Now().Add(r.lifetime))
	if err != nil {
		return RememberDeviceOut{}, fmt.Errorf("remember a device of user %s: %w", in.UserID, err)
	}

	return RememberDeviceOut{Token: token}, nil
}
