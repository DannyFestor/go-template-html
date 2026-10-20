package bootstrap

import (
	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/adapter/clock"
	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/adapter/resettoken"
	authhandler "github.com/DannyFestor/go-template-html/internal/handler/auth"
	"github.com/DannyFestor/go-template-html/internal/web"
)

const (
	forgotPasswordLimitName      = "forgot-password"
	forgotPasswordEmailLimitName = "forgot-password-email"
	resetPasswordLimitName       = "reset-password"
)

func (a *WebApp) newForgotPassword(config Config, services authServices) *authhandler.ForgotPassword {
	issue := auth.NewIssuePasswordResetLink(
		resettoken.Generator{},
		postgres.NewPasswordResetTokenStore(a.pool),
		web.NewLinks(config.App.URL),
		clock.System{},
		config.PasswordReset.TTL,
	)

	return authhandler.NewForgotPassword(
		config.App.Name,
		a.Responder,
		auth.NewRequestPasswordReset(
			a.Users,
			postgres.NewTransactor(a.pool),
			issue,
			newResetPasswordMailer(a.Jobs, config.PasswordReset.TTL),
		),
		web.NewRateLimit(services.limiter, webLimit(forgotPasswordLimitName, config.RateLimits.ForgotPassword)),
		web.NewRateLimit(
			services.limiter,
			webLimit(forgotPasswordEmailLimitName, config.RateLimits.ForgotPasswordEmail),
		),
	)
}

func (a *WebApp) newResetPassword(config Config, services authServices) *authhandler.ResetPassword {
	return authhandler.NewResetPassword(
		config.App.Name,
		a.Responder,
		auth.NewResetPassword(
			postgres.NewTransactor(a.pool),
			postgres.NewPasswordResetTokenStore(a.pool),
			auth.NewChangePassword(a.Passwords, a.Users, a.RememberTokens),
			clock.System{},
		),
		web.NewRateLimit(services.limiter, webLimit(resetPasswordLimitName, config.RateLimits.ResetPassword)),
	)
}
