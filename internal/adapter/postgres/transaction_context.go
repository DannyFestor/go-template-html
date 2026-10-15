package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type transactionKey struct{}

func withTransaction(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, transactionKey{}, tx)
}

//nolint:ireturn // pgx hands out every transaction as the pgx.Tx interface.
func transactionFrom(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(transactionKey{}).(pgx.Tx)

	return tx, ok
}
