package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
)

type executor interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

//nolint:ireturn // Inside WithinTransaction a query runs on its transaction, outside it on the pool.
func (p *Pool) executor(ctx context.Context) executor {
	if tx, ok := transactionFrom(ctx); ok {
		return tx
	}

	return p.pool
}
