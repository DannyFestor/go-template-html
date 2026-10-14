package bootstrap

import "time"

type App struct {
	Name            string        `env:"APP_NAME"         envDefault:"Go Template"`
	Env             Environment   `env:"APP_ENV,required"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"15s"`
}
