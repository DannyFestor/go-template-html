package auth

import "context"

// Broken: DisableOut is not declared.
type (
	Disable   struct{}
	DisableIn struct{}
)

func NewDisable() *Disable { return &Disable{} }

func (*Disable) Run(ctx context.Context, in DisableIn) (DisableOut, error) { return DisableOut{}, nil }
