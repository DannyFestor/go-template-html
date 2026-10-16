package postgres

import (
	"context"

	"github.com/DannyFestor/go-template-html/internal/adapter/postgres/internal/sqlcgen"
)

//nolint:ireturn // Inside WithinTransaction a query runs on its transaction, outside it on the pool.
func (p *Pool) executor(ctx context.Context) sqlcgen.DBTX {
	if tx, ok := transactionFrom(ctx); ok {
		return tx
	}

	return p.pool
}

func (p *Pool) queries(ctx context.Context) *sqlcgen.Queries {
	return sqlcgen.New(p.executor(ctx))
}
