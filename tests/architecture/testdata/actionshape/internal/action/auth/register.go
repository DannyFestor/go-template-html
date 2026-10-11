package auth

import "context"

type userCreator interface {
	Create(ctx context.Context, email string) error
}

type Register struct {
	users userCreator
}

type RegisterIn struct {
	Email string
}

type RegisterOut struct{}

func NewRegister(users userCreator) *Register {
	return &Register{users: users}
}

func (r *Register) Run(ctx context.Context, in RegisterIn) (RegisterOut, error) {
	return RegisterOut{}, r.create(ctx, in)
}

func (r *Register) create(ctx context.Context, in RegisterIn) error {
	return r.users.Create(ctx, in.Email)
}
