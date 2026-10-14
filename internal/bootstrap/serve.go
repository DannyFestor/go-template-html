package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

func Serve(ctx context.Context, server *http.Server, listener net.Listener, shutdownTimeout time.Duration) error {
	return serveUntilDone(
		ctx,
		shutdownTimeout,
		func() error { return server.Serve(listener) },
		shutdownOrClose(server),
	)
}

func shutdownOrClose(server *http.Server) func(context.Context) error {
	return func(ctx context.Context) error {
		err := server.Shutdown(ctx)
		if err != nil {
			return errors.Join(err, server.Close())
		}

		return nil
	}
}

func serveUntilDone(
	ctx context.Context,
	shutdownTimeout time.Duration,
	serve func() error,
	shutdown func(context.Context) error,
) error {
	failed := make(chan error, 1)

	go func() { failed <- serve() }()

	select {
	case err := <-failed:
		return fmt.Errorf("serve: %w", err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()

	err := shutdown(shutdownCtx)
	if err != nil {
		return fmt.Errorf("shut down within %s: %w", shutdownTimeout, err)
	}

	return nil
}
