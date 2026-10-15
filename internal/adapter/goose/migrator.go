package goose

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"path"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

type Status struct {
	Name      string
	Applied   bool
	AppliedAt time.Time
}

type Migrator struct {
	provider *goose.Provider
}

func New(db *sql.DB, migrations fs.FS) (*Migrator, error) {
	// The session lock keeps two releases that migrate at once from applying the same migration twice.
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return nil, fmt.Errorf("create the migration lock: %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations, goose.WithSessionLocker(locker))
	if err != nil {
		return nil, fmt.Errorf("load the migrations: %w", err)
	}

	return &Migrator{provider: provider}, nil
}

func (m *Migrator) Up(ctx context.Context) ([]string, error) {
	results, err := m.provider.Up(ctx)
	if err != nil {
		return nil, fmt.Errorf("apply the migrations: %w", err)
	}

	applied := make([]string, 0, len(results))
	for _, result := range results {
		applied = append(applied, path.Base(result.Source.Path))
	}

	return applied, nil
}

func (m *Migrator) Status(ctx context.Context) ([]Status, error) {
	results, err := m.provider.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("read the migration status: %w", err)
	}

	statuses := make([]Status, 0, len(results))
	for _, result := range results {
		statuses = append(statuses, Status{
			Name:      path.Base(result.Source.Path),
			Applied:   result.State == goose.StateApplied,
			AppliedAt: result.AppliedAt,
		})
	}

	return statuses, nil
}
