package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/middleware"
	"github.com/DannyFestor/go-template-html/internal/request"
	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/validate"
	"github.com/DannyFestor/go-template-html/internal/view/forgotpassword"
	"github.com/DannyFestor/go-template-html/internal/view/layout"
	"github.com/DannyFestor/go-template-html/internal/web"
)

const resetLinkSentMessage = "If an account exists, a reset link has been sent."

type passwordResetRequester interface {
	Run(ctx context.Context, in auth.RequestPasswordResetIn) (auth.RequestPasswordResetOut, error)
}

type ForgotPassword struct {
	appName     string
	responder   *web.Responder
	request     passwordResetRequester
	clientLimit rateLimit
	emailLimit  rateLimit
}

func NewForgotPassword(
	appName string,
	responder *web.Responder,
	request passwordResetRequester,
	clientLimit rateLimit,
	emailLimit rateLimit,
) *ForgotPassword {
	return &ForgotPassword{
		appName:     appName,
		responder:   responder,
		request:     request,
		clientLimit: clientLimit,
		emailLimit:  emailLimit,
	}
}

func (h *ForgotPassword) Show(w http.ResponseWriter, r *http.Request) error {
	return h.render(w, r, http.StatusOK, forgotpassword.Form{Email: "", Errors: nil, Alert: ""})
}

func (h *ForgotPassword) Store(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return fmt.Errorf("parse the forgot password form: %w", err)
	}

	forgotten := request.ParseForgotPassword(r.PostForm)

	err = h.clientLimit.Hit(r.Context(), middleware.ClientIPFrom(r.Context()).String())
	if err != nil {
		return h.refuse(w, r, err, forgotten)
	}

	invalid := forgotten.Validate()
	if len(invalid) > 0 {
		return h.render(w, r, http.StatusUnprocessableEntity, refilledForgotPassword(forgotten, invalid, ""))
	}

	err = h.requestLinkWithinTheEmailLimit(r.Context(), forgotten.Email)
	if err != nil {
		return err
	}

	h.responder.RedirectWithStatus(w, r, route.ForgotPassword, resetLinkSentMessage)

	return nil
}

func (h *ForgotPassword) requestLinkWithinTheEmailLimit(ctx context.Context, email string) error {
	err := h.emailLimit.Hit(ctx, strings.ToLower(email))
	if web.IsTooManyRequests(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("count the reset requests for the email: %w", err)
	}

	_, err = h.request.Run(ctx, auth.RequestPasswordResetIn{Email: email})
	if err != nil {
		return fmt.Errorf("request a password reset: %w", err)
	}

	return nil
}

func (h *ForgotPassword) refuse(
	w http.ResponseWriter,
	r *http.Request,
	refusal error,
	forgotten request.ForgotPassword,
) error {
	err := h.responder.RateLimited(w, r, refusal, h.page(refilledForgotPassword(forgotten, nil, rateLimitedMessage)))
	if err != nil {
		return fmt.Errorf("refuse the password reset request: %w", err)
	}

	return nil
}

func (h *ForgotPassword) render(w http.ResponseWriter, r *http.Request, status int, form forgotpassword.Form) error {
	err := h.responder.Page(w, r, status, h.page(form))
	if err != nil {
		return fmt.Errorf("show the forgot password page: %w", err)
	}

	return nil
}

func (h *ForgotPassword) page(form forgotpassword.Form) web.Page {
	return web.Page{Layout: layout.KindPublic, Body: forgotpassword.Page(h.appName, form), Fragment: nil}
}

func refilledForgotPassword(
	forgotten request.ForgotPassword,
	invalid validate.Errors,
	alert string,
) forgotpassword.Form {
	return forgotpassword.Form{Email: forgotten.Email, Errors: invalid, Alert: alert}
}
