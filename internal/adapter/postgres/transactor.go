package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Transactor struct {
	pool *pgxpool.Pool
}

func NewTransactor(pool *Pool) *Transactor {
	return &Transactor{pool: pool.pool}
}

func (t *Transactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := transactionFrom(ctx); ok {
		return fn(ctx)
	}

	err := pgx.BeginFunc(ctx, t.pool, func(tx pgx.Tx) error {
		return fn(withTransaction(ctx, tx))
	})
	if err != nil {
		return fmt.Errorf("run a transaction: %w", err)
	}

	return nil
}
