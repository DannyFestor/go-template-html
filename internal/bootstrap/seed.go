package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/seed"
)

func RunSeed() int {
	return run(Seed)
}

func Seed(ctx context.Context, config Config, logger *slog.Logger) error {
	passwords, err := newPasswordHasher(config.PasswordHashing)
	if err != nil {
		return err
	}

	pool, err := postgres.Open(ctx, postgresConfig(config.DB))
	if err != nil {
		return fmt.Errorf("connect to the database: %w", err)
	}

	defer pool.Close()

	err = seed.New(newUserStore(pool), passwords).Run(ctx)
	if err != nil {
		return fmt.Errorf("seed the database: %w", err)
	}

	logger.InfoContext(ctx, "database seeded")

	return nil
}
