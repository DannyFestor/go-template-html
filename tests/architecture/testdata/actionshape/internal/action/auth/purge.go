package auth

import "context"

// Broken: the action is not a struct.
type (
	Purge    func()
	PurgeIn  struct{}
	PurgeOut struct{}
)

func NewPurge() *Purge { return new(Purge) }

func (Purge) Run(ctx context.Context, in PurgeIn) (PurgeOut, error) { return PurgeOut{}, nil }
