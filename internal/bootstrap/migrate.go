package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"text/tabwriter"
	"time"

	"github.com/DannyFestor/go-template-html/internal/adapter/goose"
	"github.com/DannyFestor/go-template-html/internal/migrationsum"
	"github.com/DannyFestor/go-template-html/migrations"
)

var errUnknownMigrateCommand = errors.New("unknown command, want up or status")

type migrator interface {
	Up(ctx context.Context) ([]string, error)
	Status(ctx context.Context) ([]goose.Status, error)
}

type migrateCommand func(ctx context.Context, migrator migrator, logger *slog.Logger) error

func RunMigrate(args []string) int {
	command, err := prepareMigrate(args)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		return exitFailure
	}

	return run(func(ctx context.Context, config Config, logger *slog.Logger) error {
		return goose.WithMigrator(ctx, postgresConfig(config.DB), migrations.FS, func(migrator *goose.Migrator) error {
			return command(ctx, migrator, logger)
		})
	})
}

// The sum is checked before connecting, so a binary with tampered migrations never touches the database.
func prepareMigrate(args []string) (migrateCommand, error) {
	command, err := parseMigrateCommand(args)
	if err != nil {
		return nil, err
	}

	err = migrationsum.Verify(migrations.FS)
	if err != nil {
		return nil, fmt.Errorf("verify the migrations: %w", err)
	}

	return command, nil
}

func parseMigrateCommand(args []string) (migrateCommand, error) {
	commands := map[string]migrateCommand{
		"up":     migrateUp,
		"status": func(ctx context.Context, m migrator, _ *slog.Logger) error { return printStatus(ctx, m, os.Stdout) },
	}

	if len(args) == 1 && commands[args[0]] != nil {
		return commands[args[0]], nil
	}

	return nil, fmt.Errorf("migrate %q: %w", args, errUnknownMigrateCommand)
}

func migrateUp(ctx context.Context, migrator migrator, logger *slog.Logger) error {
	applied, err := migrator.Up(ctx)
	if err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}

	for _, name := range applied {
		logger.InfoContext(ctx, "migration applied", slog.String("migration", name))
	}

	logger.InfoContext(ctx, "migrations up to date", slog.Int("applied", len(applied)))

	return nil
}

func printStatus(ctx context.Context, migrator migrator, out io.Writer) error {
	statuses, err := migrator.Status(ctx)
	if err != nil {
		return fmt.Errorf("migrate status: %w", err)
	}

	table := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0) //nolint:mnd // Two spaces between the columns.
	for _, status := range statuses {
		_, _ = fmt.Fprintf(table, "%s\t%s\t%s\n", statusState(status), statusAppliedAt(status), status.Name)
	}

	err = table.Flush()
	if err != nil {
		return fmt.Errorf("print the migration status: %w", err)
	}

	return nil
}

func statusState(status goose.Status) string {
	if status.Applied {
		return "applied"
	}

	return "pending"
}

func statusAppliedAt(status goose.Status) string {
	if !status.Applied {
		return "-"
	}

	return status.AppliedAt.Format(time.DateTime)
}
