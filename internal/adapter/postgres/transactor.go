package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Transactor struct {
	pool *Pool
}

func NewTransactor(pool *Pool) *Transactor {
	return &Transactor{pool: pool}
}

func (t *Transactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if inTransaction(ctx) {
		return fn(ctx)
	}

	err := pgx.BeginFunc(ctx, t.pool.pool, func(tx pgx.Tx) error {
		return fn(withTransaction(ctx, tx))
	})
	if err != nil {
		return fmt.Errorf("within a transaction: %w", err)
	}

	return nil
}
