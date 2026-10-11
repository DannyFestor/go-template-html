package auth

import "context"

// Broken: the constructor returns a value, not a pointer.
type (
	Verify    struct{}
	VerifyIn  struct{}
	VerifyOut struct{}
)

func NewVerify() Verify { return Verify{} }

func (*Verify) Run(ctx context.Context, in VerifyIn) (VerifyOut, error) { return VerifyOut{}, nil }
