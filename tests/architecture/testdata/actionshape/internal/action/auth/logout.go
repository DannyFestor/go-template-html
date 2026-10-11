package auth

import "context"

// Broken: no NewLogout constructor.
type (
	Logout    struct{}
	LogoutIn  struct{}
	LogoutOut struct{}
)

func (*Logout) Run(ctx context.Context, in LogoutIn) (LogoutOut, error) { return LogoutOut{}, nil }
