package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

var ErrIncorrectPassword = errors.New("the password is not the user's current password")

type ConfirmPassword struct {
	passwords passwordVerifier
}

type ConfirmPasswordIn struct {
	User     domain.User
	Password string
}

type ConfirmPasswordOut struct{}

func NewConfirmPassword(passwords passwordVerifier) *ConfirmPassword {
	return &ConfirmPassword{passwords: passwords}
}

func (c *ConfirmPassword) Run(_ context.Context, in ConfirmPasswordIn) (ConfirmPasswordOut, error) {
	matches, err := c.passwords.Verify(in.Password, string(in.User.PasswordHash))
	if err != nil {
		return ConfirmPasswordOut{}, fmt.Errorf("confirm the password of user %s: %w", in.User.ID, err)
	}

	if !matches {
		return ConfirmPasswordOut{}, ErrIncorrectPassword
	}

	return ConfirmPasswordOut{}, nil
}
