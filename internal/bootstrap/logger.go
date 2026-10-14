package bootstrap

import (
	"io"
	"log/slog"

	"github.com/DannyFestor/go-template-html/internal/middleware"
)

func NewLogger(out io.Writer, environment Environment, config Log) *slog.Logger {
	options := &slog.HandlerOptions{AddSource: false, Level: config.Level, ReplaceAttr: nil}

	return slog.New(middleware.NewRequestIDLogHandler(newLogHandler(out, environment, options)))
}

func newLogHandler(out io.Writer, environment Environment, options *slog.HandlerOptions) slog.Handler {
	if environment == EnvironmentProduction {
		return slog.NewJSONHandler(out, options)
	}

	return slog.NewTextHandler(out, options)
}
