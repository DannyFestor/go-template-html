package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

var ErrNotRemembered = errors.New("the remember token is unknown, expired or its user is gone")

type Recall struct {
	transactor transactor
	tokens     rememberTokenConsumer
	users      userFinder
	remember   deviceRememberer
	clock      clock
}

type RecallIn struct {
	Token domain.RememberToken
}

type RecallOut struct {
	User  domain.User
	Token domain.RememberToken
}

func NewRecall(
	transactor transactor,
	tokens rememberTokenConsumer,
	users userFinder,
	remember deviceRememberer,
	clock clock,
) *Recall {
	return &Recall{transactor: transactor, tokens: tokens, users: users, remember: remember, clock: clock}
}

func (r *Recall) Run(ctx context.Context, in RecallIn) (RecallOut, error) {
	var recalled RecallOut

	err := r.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		var err error

		recalled, err = r.recall(ctx, in.Token)

		return err
	})
	if err != nil {
		return RecallOut{}, fmt.Errorf("recall a remembered user: %w", err)
	}

	return recalled, nil
}

func (r *Recall) recall(ctx context.Context, token domain.RememberToken) (RecallOut, error) {
	use, err := r.consume(ctx, token)
	if err != nil {
		return RecallOut{}, err
	}

	user, err := r.findUser(ctx, use.UserID)
	if err != nil {
		return RecallOut{}, err
	}

	if use.Replay {
		var rotatedByTheFirstUse domain.RememberToken

		return RecallOut{User: user, Token: rotatedByTheFirstUse}, nil
	}

	remembered, err := r.remember.Run(ctx, RememberDeviceIn{UserID: user.ID})
	if err != nil {
		return RecallOut{}, fmt.Errorf("rotate the remember token: %w", err)
	}

	return RecallOut{User: user, Token: remembered.Token}, nil
}

func (r *Recall) consume(ctx context.Context, token domain.RememberToken) (domain.RememberTokenUse, error) {
	use, consumed, err := r.tokens.Consume(ctx, token, r.clock.Now())
	if err != nil {
		return use, fmt.Errorf("consume the remember token: %w", err)
	}

	if !consumed {
		return use, ErrNotRemembered
	}

	return use, nil
}

func (r *Recall) findUser(ctx context.Context, userID domain.UserID) (domain.User, error) {
	user, found, err := r.users.Find(ctx, userID)
	if err != nil {
		return domain.User{}, fmt.Errorf("find the remembered user: %w", err)
	}

	if !found {
		return domain.User{}, ErrNotRemembered
	}

	return user, nil
}
