package bootstrap

import (
	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/adapter/clock"
	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/adapter/remembertoken"
	"github.com/DannyFestor/go-template-html/internal/adapter/valkey"
	authhandler "github.com/DannyFestor/go-template-html/internal/handler/auth"
	"github.com/DannyFestor/go-template-html/internal/middleware"
	"github.com/DannyFestor/go-template-html/internal/web"
)

const (
	registrationLimitName = "registration"
	loginLimitName        = "login"
)

type authRoutes struct {
	register     *authhandler.Register
	login        *authhandler.Login
	logout       *authhandler.Logout
	requireUser  wrapper
	requireGuest wrapper
}

func (a *WebApp) newAuthRoutes(config Config) (authRoutes, error) {
	credentials, err := newCredentialFingerprints(config.App.Key)
	if err != nil {
		return authRoutes{register: nil, login: nil, logout: nil, requireUser: nil, requireGuest: nil}, err
	}

	authenticator := middleware.NewAuthenticator(a.Users, credentials, a.Responder)
	limiter := valkey.NewRateLimiter(a.valkey)
	rememberDevice := auth.NewRememberDevice(
		remembertoken.Generator{},
		a.RememberTokens,
		clock.System{},
		config.Session.RememberLifetime,
	)
	sessions := a.newSessions(config, authenticator, rememberDevice)

	return authRoutes{
		register: authhandler.NewRegister(
			config.App.Name,
			a.Responder,
			auth.NewRegister(a.Users, a.Passwords),
			authenticator,
			web.NewRateLimit(limiter, webLimit(registrationLimitName, config.RateLimits.Registration)),
		),
		login: authhandler.NewLogin(
			config.App.Name,
			a.Responder,
			auth.NewLogin(a.Users, a.Passwords, rememberDevice),
			sessions,
			web.NewRateLimit(limiter, webLimit(loginLimitName, config.RateLimits.Login)),
		),
		logout:       authhandler.NewLogout(a.Responder, sessions),
		requireUser:  chain(sessions.RestoreRemembered, authenticator.Require),
		requireGuest: chain(sessions.RestoreRemembered, authenticator.RequireGuest),
	}, nil
}

func (a *WebApp) newSessions(
	config Config,
	authenticator *middleware.Authenticator,
	rememberDevice *auth.RememberDevice,
) *authhandler.Sessions {
	return authhandler.NewSessions(
		a.Responder,
		auth.NewRecall(postgres.NewTransactor(a.pool), a.RememberTokens, a.Users, rememberDevice, clock.System{}),
		auth.NewLogout(a.RememberTokens),
		authenticator,
		middleware.NewRememberCookie(config.App.servedOverHTTPS(), config.Session.RememberLifetime),
	)
}

func webLimit(name string, limit RateLimit) web.Limit {
	return web.Limit{Name: name, MaxAttempts: limit.MaxAttempts, Window: limit.Window}
}
