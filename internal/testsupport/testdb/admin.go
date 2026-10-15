package testdb

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const maintenanceDatabase = "postgres"

type admin struct {
	conn *pgx.Conn
}

// Advisory locks are scoped to one database, so every test process takes its lock in the same maintenance database.
func withAdmin(ctx context.Context, s settings, use func(admin) error) error {
	conn, err := pgx.Connect(ctx, s.database(maintenanceDatabase).ConnString())
	if err != nil {
		return fmt.Errorf("connect to the %s database: %w", maintenanceDatabase, err)
	}

	defer func() { _ = conn.Close(context.WithoutCancel(ctx)) }()

	return use(admin{conn: conn})
}

func (a admin) exec(ctx context.Context, sql string, args ...any) error {
	_, err := a.conn.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", sql, err)
	}

	return nil
}

func (a admin) databaseExists(ctx context.Context, name string) (bool, error) {
	var exists bool

	err := a.conn.QueryRow(ctx, "SELECT EXISTS (SELECT FROM pg_database WHERE datname = $1)", name).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("look up database %s: %w", name, err)
	}

	return exists, nil
}

func (a admin) databasesStartingWith(ctx context.Context, prefix string) ([]string, error) {
	rows, err := a.conn.Query(ctx, "SELECT datname FROM pg_database WHERE starts_with(datname, $1)", prefix)
	if err != nil {
		return nil, fmt.Errorf("list the databases starting with %s: %w", prefix, err)
	}

	names, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, fmt.Errorf("list the databases starting with %s: %w", prefix, err)
	}

	return names, nil
}

func (a admin) create(ctx context.Context, name, template string) error {
	return a.exec(ctx, "CREATE DATABASE "+quote(name)+" TEMPLATE "+quote(template))
}

func (a admin) drop(ctx context.Context, name string) error {
	return a.exec(ctx, "DROP DATABASE IF EXISTS "+quote(name)+" WITH (FORCE)")
}

func (a admin) rename(ctx context.Context, from, to string) error {
	return a.exec(ctx, "ALTER DATABASE "+quote(from)+" RENAME TO "+quote(to))
}

func quote(name string) string {
	return pgx.Identifier{name}.Sanitize()
}
