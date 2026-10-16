package bootstrap

import "time"

type SignedURL struct {
	TTL time.Duration `env:"SIGNED_URL_TTL" envDefault:"2h"`
}
