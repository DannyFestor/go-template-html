package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/middleware"
	"github.com/DannyFestor/go-template-html/internal/request"
	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/validate"
	"github.com/DannyFestor/go-template-html/internal/view/layout"
	"github.com/DannyFestor/go-template-html/internal/view/resetpassword"
	"github.com/DannyFestor/go-template-html/internal/web"
)

const (
	passwordResetMessage    = "Your password has been reset. Log in with your new password."
	invalidResetLinkMessage = "This password reset link is invalid or has expired."
)

type passwordResetter interface {
	Run(ctx context.Context, in auth.ResetPasswordIn) (auth.ResetPasswordOut, error)
}

type ResetPassword struct {
	appName   string
	responder *web.Responder
	reset     passwordResetter
	limit     rateLimit
}

func NewResetPassword(
	appName string,
	responder *web.Responder,
	reset passwordResetter,
	limit rateLimit,
) *ResetPassword {
	return &ResetPassword{appName: appName, responder: responder, reset: reset, limit: limit}
}

func (h *ResetPassword) Show(w http.ResponseWriter, r *http.Request) error {
	token := r.URL.Query().Get(route.ResetPasswordTokenParam)

	return h.render(w, r, http.StatusOK, resetpassword.Form{Token: token, Errors: nil, Alert: ""})
}

func (h *ResetPassword) Store(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return fmt.Errorf("parse the reset password form: %w", err)
	}

	reset := request.ParsePasswordReset(r.PostForm)

	err = h.limit.Hit(r.Context(), middleware.ClientIPFrom(r.Context()).String())
	if err != nil {
		return h.refuse(w, r, err, reset)
	}

	invalid := reset.Validate()
	if len(invalid) > 0 {
		return h.render(w, r, http.StatusUnprocessableEntity, refilledPasswordReset(reset, invalid, ""))
	}

	return h.resetPassword(w, r, reset)
}

func (h *ResetPassword) resetPassword(w http.ResponseWriter, r *http.Request, reset request.PasswordReset) error {
	_, err := h.reset.Run(r.Context(), auth.ResetPasswordIn{
		Token:    domain.PasswordResetToken(reset.Token),
		Password: reset.Password,
	})
	if errors.Is(err, auth.ErrInvalidPasswordResetToken) {
		return h.render(
			w,
			r,
			http.StatusUnprocessableEntity,
			refilledPasswordReset(reset, nil, invalidResetLinkMessage),
		)
	}

	if err != nil {
		return fmt.Errorf("reset the password: %w", err)
	}

	h.responder.RedirectWithStatus(w, r, route.Login, passwordResetMessage)

	return nil
}

func (h *ResetPassword) refuse(
	w http.ResponseWriter,
	r *http.Request,
	refusal error,
	reset request.PasswordReset,
) error {
	web.SkipCompression(w)

	err := h.responder.RateLimited(w, r, refusal, h.page(refilledPasswordReset(reset, nil, rateLimitedMessage)))
	if err != nil {
		return fmt.Errorf("refuse the password reset: %w", err)
	}

	return nil
}

func (h *ResetPassword) render(w http.ResponseWriter, r *http.Request, status int, form resetpassword.Form) error {
	web.SkipCompression(w)

	err := h.responder.Page(w, r, status, h.page(form))
	if err != nil {
		return fmt.Errorf("show the reset password page: %w", err)
	}

	return nil
}

func (h *ResetPassword) page(form resetpassword.Form) web.Page {
	return web.Page{Layout: layout.KindPublic, Body: resetpassword.Page(h.appName, form), Fragment: nil}
}

func refilledPasswordReset(reset request.PasswordReset, invalid validate.Errors, alert string) resetpassword.Form {
	return resetpassword.Form{Token: reset.Token, Errors: invalid, Alert: alert}
}
