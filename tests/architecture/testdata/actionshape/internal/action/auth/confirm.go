package auth

import "context"

// Broken: a second exported method.
type (
	Confirm    struct{}
	ConfirmIn  struct{}
	ConfirmOut struct{}
)

func NewConfirm() *Confirm { return &Confirm{} }

func (*Confirm) Run(ctx context.Context, in ConfirmIn) (ConfirmOut, error) { return ConfirmOut{}, nil }

func (*Confirm) Reset() {}
