package logreport

import (
	"context"
	"log/slog"
	"runtime/debug"
)

type Reporter struct {
	logger *slog.Logger
}

func New(logger *slog.Logger) *Reporter {
	return &Reporter{logger: logger}
}

func (r *Reporter) Report(ctx context.Context, err error) {
	r.logger.ErrorContext(
		ctx,
		"unexpected error",
		slog.String("error", err.Error()),
		slog.String("stack", string(debug.Stack())),
	)
}
