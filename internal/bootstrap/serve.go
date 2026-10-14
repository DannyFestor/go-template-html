package bootstrap

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

func Serve(ctx context.Context, server *http.Server, listener net.Listener, shutdownTimeout time.Duration) error {
	return runUntilDone(ctx, shutdownTimeout, func() error { return server.Serve(listener) }, server.Shutdown)
}

func runUntilDone(
	ctx context.Context,
	stopTimeout time.Duration,
	run func() error,
	stop func(context.Context) error,
) error {
	failed := make(chan error, 1)

	go func() { failed <- run() }()

	select {
	case err := <-failed:
		return fmt.Errorf("run: %w", err)
	case <-ctx.Done():
	}

	stopCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), stopTimeout)
	defer cancel()

	err := stop(stopCtx)
	if err != nil {
		return fmt.Errorf("stop within %s: %w", stopTimeout, err)
	}

	return nil
}
