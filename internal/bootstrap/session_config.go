package bootstrap

import "time"

type Session struct {
	IdleLifetime                time.Duration `env:"SESSION_IDLE_LIFETIME"         envDefault:"2h"`
	AbsoluteLifetime            time.Duration `env:"SESSION_ABSOLUTE_LIFETIME"     envDefault:"24h"`
	RememberLifetime            time.Duration `env:"REMEMBER_LIFETIME"             envDefault:"720h"`
	PasswordConfirmationTimeout time.Duration `env:"PASSWORD_CONFIRMATION_TIMEOUT" envDefault:"3h"`
}
