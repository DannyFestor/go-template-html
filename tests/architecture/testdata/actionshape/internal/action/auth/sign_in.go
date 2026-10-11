package auth

import "context"

type (
	SignIn    struct{}
	SignInIn  struct{}
	SignInOut struct{}
)

func NewSignIn() *SignIn { return &SignIn{} }

func (*SignIn) Run(ctx context.Context, in SignInIn) (SignInOut, error) { return SignInOut{}, nil }
