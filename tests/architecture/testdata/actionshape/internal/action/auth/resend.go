package auth

import "context"

// Broken: ResendIn is not a struct.
type (
	Resend    struct{}
	ResendIn  string
	ResendOut struct{}
)

func NewResend() *Resend { return &Resend{} }

func (*Resend) Run(ctx context.Context, in ResendIn) (ResendOut, error) { return ResendOut{}, nil }
