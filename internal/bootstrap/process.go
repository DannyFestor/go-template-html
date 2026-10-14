package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/DannyFestor/go-template-html/internal/adapter/logreport"
)

const (
	dotenvPath  = ".env"
	exitSuccess = 0
	exitFailure = 1
)

type errorReporter interface {
	Report(ctx context.Context, err error)
}

type process func(ctx context.Context, config Config, logger *slog.Logger) error

func run(serve process) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	// Restoring the default handling after the first signal lets a second one kill a drain that hangs.
	context.AfterFunc(ctx, stop)

	config, err := LoadConfig(dotenvPath, os.Environ())
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		return exitFailure
	}

	logger := NewLogger(os.Stdout, config.App.Env, config.Log)

	return exitCode(ctx, logreport.New(logger), serve(ctx, config, logger))
}

func exitCode(ctx context.Context, reporter errorReporter, err error) int {
	if err == nil {
		return exitSuccess
	}

	reporter.Report(ctx, err)

	return exitFailure
}
