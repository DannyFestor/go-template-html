package postgres

import (
	"context"
	"fmt"
)

func Exec(ctx context.Context, pool *Pool, sql string, args ...any) error {
	_, err := pool.executor(ctx).Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", sql, err)
	}

	return nil
}
