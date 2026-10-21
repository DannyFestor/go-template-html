package bootstrap

import (
	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/adapter/appkey"
	"github.com/DannyFestor/go-template-html/internal/adapter/clock"
	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/adapter/remembertoken"
	"github.com/DannyFestor/go-template-html/internal/adapter/signedurl"
	"github.com/DannyFestor/go-template-html/internal/adapter/valkey"
	"github.com/DannyFestor/go-template-html/internal/adapter/verificationlink"
	authhandler "github.com/DannyFestor/go-template-html/internal/handler/auth"
	"github.com/DannyFestor/go-template-html/internal/middleware"
	"github.com/DannyFestor/go-template-html/internal/web"
)

const (
	registrationLimitName         = "registration"
	loginLimitName                = "login"
	passwordConfirmationLimitName = "password-confirmation"
	verificationResendLimitName   = "verification-resend"
	verificationLinkLimitName     = "verification-link"
)

type authRoutes struct {
	register                    *authhandler.Register
	login                       *authhandler.Login
	logout                      *authhandler.Logout
	confirmPassword             *authhandler.ConfirmPassword
	verificationNotice          *authhandler.VerificationNotice
	verifyEmail                 *authhandler.VerifyEmail
	forgotPassword              *authhandler.ForgotPassword
	resetPassword               *authhandler.ResetPassword
	requireUser                 wrapper
	requireVerified             wrapper
	requireSignedLink           wrapper
	requireGuest                wrapper
	requirePasswordConfirmation wrapper
}

type authServices struct {
	authenticator    *middleware.Authenticator
	limiter          *valkey.RateLimiter
	rememberDevice   *auth.RememberDevice
	sessions         *authhandler.Sessions
	confirmation     *middleware.PasswordConfirmation
	links            *verificationlink.Links
	sendVerification *auth.SendVerificationMail
	signatures       *signedurl.Verifier
}

func (a *WebApp) newAuthRoutes(config Config) (authRoutes, error) {
	var none authRoutes

	services, err := a.newAuthServices(config)
	if err != nil {
		return none, err
	}

	requireUser := chain(services.sessions.RestoreRemembered, services.authenticator.Require)

	return authRoutes{
		register:           a.newRegister(config, services),
		login:              a.newLogin(config, services),
		logout:             authhandler.NewLogout(a.Responder, services.sessions),
		confirmPassword:    a.newConfirmPassword(config, services),
		verificationNotice: a.newVerificationNotice(config, services),
		verifyEmail:        a.newVerifyEmail(config, services),
		forgotPassword:     a.newForgotPassword(config, services),
		resetPassword:      a.newResetPassword(config, services),
		requireUser:        requireUser,
		requireVerified:    chain(requireUser, services.authenticator.RequireVerified),
		requireSignedLink: chain(
			requireUser,
			middleware.RequireValidSignature(services.signatures, a.Responder),
		),
		requireGuest:                chain(services.sessions.RestoreRemembered, services.authenticator.RequireGuest),
		requirePasswordConfirmation: chain(requireUser, services.confirmation.Require),
	}, nil
}

func (a *WebApp) newAuthServices(config Config) (authServices, error) {
	var none authServices

	credentials, err := newCredentialFingerprints(config.App.Key)
	if err != nil {
		return none, err
	}

	signedURLKey, err := subkeyFor(config.App.Key, appkey.PurposeSignedURL)
	if err != nil {
		return none, err
	}

	authenticator := middleware.NewAuthenticator(a.Users, credentials, a.Responder)
	rememberDevice := a.newRememberDevice(config)
	links := newVerificationLinks(config, signedURLKey)

	return authServices{
		authenticator:    authenticator,
		limiter:          valkey.NewRateLimiter(a.valkey),
		rememberDevice:   rememberDevice,
		sessions:         a.newSessions(config, authenticator, rememberDevice),
		confirmation:     a.newPasswordConfirmation(config),
		links:            links,
		sendVerification: auth.NewSendVerificationMail(links, newVerificationMailer(a.Jobs)),
		signatures:       signedurl.NewVerifier(signedURLKey, clock.System{}),
	}, nil
}

func (a *WebApp) newRegister(config Config, services authServices) *authhandler.Register {
	return authhandler.NewRegister(
		config.App.Name,
		a.Responder,
		auth.NewRegister(postgres.NewTransactor(a.pool), a.Users, a.Passwords, services.sendVerification),
		services.authenticator,
		web.NewRateLimit(services.limiter, webLimit(registrationLimitName, config.RateLimits.Registration)),
	)
}

func (a *WebApp) newLogin(config Config, services authServices) *authhandler.Login {
	return authhandler.NewLogin(
		config.App.Name,
		a.Responder,
		auth.NewLogin(a.Users, a.Passwords, services.rememberDevice),
		services.sessions,
		web.NewRateLimit(services.limiter, webLimit(loginLimitName, config.RateLimits.Login)),
	)
}

func (a *WebApp) newConfirmPassword(config Config, services authServices) *authhandler.ConfirmPassword {
	return authhandler.NewConfirmPassword(
		config.App.Name,
		a.Responder,
		auth.NewConfirmPassword(a.Passwords),
		services.confirmation,
		web.NewRateLimit(
			services.limiter,
			webLimit(passwordConfirmationLimitName, config.RateLimits.PasswordConfirmation),
		),
	)
}

func (a *WebApp) newVerificationNotice(config Config, services authServices) *authhandler.VerificationNotice {
	return authhandler.NewVerificationNotice(
		config.App.Name,
		a.Responder,
		services.sendVerification,
		web.NewRateLimit(services.limiter, webLimit(verificationResendLimitName, config.RateLimits.VerificationResend)),
	)
}

func (a *WebApp) newVerifyEmail(config Config, services authServices) *authhandler.VerifyEmail {
	return authhandler.NewVerifyEmail(
		a.Responder,
		auth.NewVerifyEmail(services.links, a.Users, clock.System{}),
		services.authenticator,
		web.NewRateLimit(services.limiter, webLimit(verificationLinkLimitName, config.RateLimits.VerificationLink)),
	)
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

func webLimit(name string, limit RateLimit) web.Limit {
	return web.Limit{Name: name, MaxAttempts: limit.MaxAttempts, Window: limit.Window}
}
