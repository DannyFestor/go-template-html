package bootstrap

import (
	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/adapter/valkey"
	authhandler "github.com/DannyFestor/go-template-html/internal/handler/auth"
	"github.com/DannyFestor/go-template-html/internal/middleware"
	"github.com/DannyFestor/go-template-html/internal/web"
)

const registrationLimitName = "registration"

type authRoutes struct {
	register     *authhandler.Register
	requireUser  wrapper
	requireGuest wrapper
}

func (a *WebApp) newAuthRoutes(config Config) (authRoutes, error) {
	credentials, err := newCredentialFingerprints(config.App.Key)
	if err != nil {
		return authRoutes{register: nil, requireUser: nil, requireGuest: nil}, err
	}

	authenticator := middleware.NewAuthenticator(a.Users, credentials, a.Responder)
	limiter := valkey.NewRateLimiter(a.valkey)

	return authRoutes{
		register: authhandler.NewRegister(
			config.App.Name,
			a.Responder,
			auth.NewRegister(a.Users, a.Passwords),
			authenticator,
			web.NewRateLimit(limiter, webLimit(registrationLimitName, config.RateLimits.Registration)),
		),
		requireUser:  authenticator.Require,
		requireGuest: authenticator.RequireGuest,
	}, nil
}

func webLimit(name string, limit RateLimit) web.Limit {
	return web.Limit{Name: name, MaxAttempts: limit.MaxAttempts, Window: limit.Window}
}
