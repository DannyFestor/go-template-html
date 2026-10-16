package bootstrap

import (
	"net/url"
	"time"
)

type App struct {
	Name            string        `env:"APP_NAME"         envDefault:"Go Template"`
	Env             Environment   `env:"APP_ENV,required"`
	URL             url.URL       `env:"APP_URL,required"`
	Key             AppKey        `env:"APP_KEY,required"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"15s"`
}

func (a App) servedOverHTTPS() bool {
	return a.URL.Scheme == "https"
}
