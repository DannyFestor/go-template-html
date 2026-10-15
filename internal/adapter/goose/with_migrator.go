package goose

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
)

func WithMigrator(ctx context.Context, config postgres.Config, migrations fs.FS, use func(*Migrator) error) error {
	pool, err := postgres.Open(ctx, config)
	if err != nil {
		return fmt.Errorf("connect to the database: %w", err)
	}
	defer pool.Close()

	err = pool.EnsureSchema(ctx)
	if err != nil {
		return fmt.Errorf("prepare the schema: %w", err)
	}

	db := pool.OpenSQLDB()
	defer func() { _ = db.Close() }()

	migrator, err := New(db, migrations)
	if err != nil {
		return err
	}

	return use(migrator)
}
