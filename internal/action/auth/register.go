package auth

import (
	"context"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type Register struct {
	transactor   transactor
	users        userCreator
	passwords    passwordHasher
	verification verificationSender
}

type RegisterIn struct {
	Name     string
	Email    string
	Password string
}

type RegisterOut struct {
	User domain.User
}

func NewRegister(
	transactor transactor,
	users userCreator,
	passwords passwordHasher,
	verification verificationSender,
) *Register {
	return &Register{transactor: transactor, users: users, passwords: passwords, verification: verification}
}

func (r *Register) Run(ctx context.Context, in RegisterIn) (RegisterOut, error) {
	var registered RegisterOut

	err := r.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		var err error

		registered, err = r.register(ctx, in)

		return err
	})
	if err != nil {
		return RegisterOut{}, fmt.Errorf("register %s: %w", in.Email, err)
	}

	return registered, nil
}

func (r *Register) register(ctx context.Context, in RegisterIn) (RegisterOut, error) {
	user, err := r.users.Create(ctx, domain.NewUser{
		Name:          in.Name,
		Email:         in.Email,
		PasswordHash:  domain.PasswordHash(r.passwords.Hash(in.Password)),
		EmailVerified: false,
	})
	if err != nil {
		return RegisterOut{}, fmt.Errorf("create the user: %w", err)
	}

	_, err = r.verification.Run(ctx, SendVerificationMailIn{User: user})
	if err != nil {
		return RegisterOut{}, fmt.Errorf("mail the verification link: %w", err)
	}

	return RegisterOut{User: user}, nil
}
