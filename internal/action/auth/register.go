package auth

import (
	"context"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type Register struct {
	users     userCreator
	passwords passwordHasher
}

type RegisterIn struct {
	Name     string
	Email    string
	Password string
}

type RegisterOut struct {
	User domain.User
}

func NewRegister(users userCreator, passwords passwordHasher) *Register {
	return &Register{users: users, passwords: passwords}
}

func (r *Register) Run(ctx context.Context, in RegisterIn) (RegisterOut, error) {
	user, err := r.users.Create(ctx, domain.NewUser{
		Name:         in.Name,
		Email:        in.Email,
		PasswordHash: domain.PasswordHash(r.passwords.Hash(in.Password)),
	})
	if err != nil {
		return RegisterOut{}, fmt.Errorf("register %s: %w", in.Email, err)
	}

	return RegisterOut{User: user}, nil
}
