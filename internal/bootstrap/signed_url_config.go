package bootstrap

import "time"

type SignedURL struct {
	TTL                  time.Duration `env:"SIGNED_URL_TTL"         envDefault:"2h"`
	EmailVerificationTTL time.Duration `env:"EMAIL_VERIFICATION_TTL" envDefault:"60m"`
}
