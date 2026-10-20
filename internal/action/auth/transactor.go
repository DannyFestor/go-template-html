package auth

import "context"

//go:generate go tool mockgen -source=transactor.go -destination=mock_transactor_test.go -package=auth -mock_names=transactor=MockTransactor

type transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
