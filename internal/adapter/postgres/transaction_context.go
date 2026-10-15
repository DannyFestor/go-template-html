package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type transactionKey struct{}

func withTransaction(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, transactionKey{}, tx)
}

func inTransaction(ctx context.Context) bool {
	return ctx.Value(transactionKey{}) != nil
}
