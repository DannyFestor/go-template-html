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
	registrationLimitName         = "registration"
	loginLimitName                = "login"
	passwordConfirmationLimitName = "password-confirmation"
)

type authRoutes struct {
	register                    *authhandler.Register
	login                       *authhandler.Login
	logout                      *authhandler.Logout
	confirmPassword             *authhandler.ConfirmPassword
	requireUser                 wrapper
	requireGuest                wrapper
	requirePasswordConfirmation wrapper
}

func (a *WebApp) newAuthRoutes(config Config) (authRoutes, error) {
	var none authRoutes

	credentials, err := newCredentialFingerprints(config.App.Key)
	if err != nil {
		return none, err
	}

	authenticator := middleware.NewAuthenticator(a.Users, credentials, a.Responder)
	limiter := valkey.NewRateLimiter(a.valkey)
	rememberDevice := a.newRememberDevice(config)
	sessions := a.newSessions(config, authenticator, rememberDevice)
	confirmation := a.newPasswordConfirmation(config)
	requireUser := chain(sessions.RestoreRemembered, authenticator.Require)

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
		logout:                      authhandler.NewLogout(a.Responder, sessions),
		confirmPassword:             a.newConfirmPassword(config, limiter, confirmation),
		requireUser:                 requireUser,
		requireGuest:                chain(sessions.RestoreRemembered, authenticator.RequireGuest),
		requirePasswordConfirmation: chain(requireUser, confirmation.Require),
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

func (a *WebApp) newRememberDevice(config Config) *auth.RememberDevice {
	return auth.NewRememberDevice(
		remembertoken.Generator{},
		a.RememberTokens,
		clock.System{},
		config.Session.RememberLifetime,
	)
}

func (a *WebApp) newPasswordConfirmation(config Config) *middleware.PasswordConfirmation {
	return middleware.NewPasswordConfirmation(clock.System{}, config.Session.PasswordConfirmationTimeout, a.Responder)
}

func (a *WebApp) newConfirmPassword(
	config Config,
	limiter *valkey.RateLimiter,
	confirmation *middleware.PasswordConfirmation,
) *authhandler.ConfirmPassword {
	return authhandler.NewConfirmPassword(
		config.App.Name,
		a.Responder,
		auth.NewConfirmPassword(a.Passwords),
		confirmation,
		web.NewRateLimit(limiter, webLimit(passwordConfirmationLimitName, config.RateLimits.PasswordConfirmation)),
	)
}

func webLimit(name string, limit RateLimit) web.Limit {
	return web.Limit{Name: name, MaxAttempts: limit.MaxAttempts, Window: limit.Window}
}
