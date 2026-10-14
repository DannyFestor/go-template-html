package bootstrap

import (
	"io"
	"log/slog"
)

func NewLogger(out io.Writer, environment Environment, config Log) *slog.Logger {
	options := &slog.HandlerOptions{AddSource: false, Level: config.Level, ReplaceAttr: nil}

	if environment == EnvironmentProduction {
		return slog.New(slog.NewJSONHandler(out, options))
	}

	return slog.New(slog.NewTextHandler(out, options))
}
