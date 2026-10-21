package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

var ErrInvalidPasswordResetToken = errors.New("the password reset token is unknown, used or expired")

type ResetPassword struct {
	transactor transactor
	tokens     passwordResetTokenConsumer
	password   passwordChanger
	clock      clock
}

type ResetPasswordIn struct {
	Token    domain.PasswordResetToken
	Password string
}

type ResetPasswordOut struct{}

func NewResetPassword(
	transactor transactor,
	tokens passwordResetTokenConsumer,
	password passwordChanger,
	clock clock,
) *ResetPassword {
	return &ResetPassword{transactor: transactor, tokens: tokens, password: password, clock: clock}
}

func (r *ResetPassword) Run(ctx context.Context, in ResetPasswordIn) (ResetPasswordOut, error) {
	err := r.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		return r.reset(ctx, in)
	})
	if err != nil {
		return ResetPasswordOut{}, fmt.Errorf("reset a password: %w", err)
	}

	return ResetPasswordOut{}, nil
}

func (r *ResetPassword) reset(ctx context.Context, in ResetPasswordIn) error {
	userID, err := r.consume(ctx, in.Token)
	if err != nil {
		return err
	}

	_, err = r.password.Run(ctx, ChangePasswordIn{UserID: userID, Password: in.Password})
	if err != nil {
		return fmt.Errorf("change the password: %w", err)
	}

	return nil
}

func (r *ResetPassword) consume(ctx context.Context, token domain.PasswordResetToken) (domain.UserID, error) {
	userID, consumed, err := r.tokens.Consume(ctx, token, r.clock.Now())
	if err != nil {
		return userID, fmt.Errorf("consume the reset token: %w", err)
	}

	if !consumed {
		return userID, ErrInvalidPasswordResetToken
	}

	return userID, nil
}
