package bootstrap

import "log/slog"

type Log struct {
	Level slog.Level `env:"LOG_LEVEL" envDefault:"info"`
}
