package bootstrap

import "time"

type PasswordReset struct {
	TTL time.Duration `env:"PASSWORD_RESET_TTL" envDefault:"60m"`
}
