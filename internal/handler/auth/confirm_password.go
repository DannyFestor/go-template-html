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
	"github.com/DannyFestor/go-template-html/internal/validate"
	"github.com/DannyFestor/go-template-html/internal/view/confirmpassword"
	"github.com/DannyFestor/go-template-html/internal/view/layout"
	"github.com/DannyFestor/go-template-html/internal/web"
)

const incorrectPasswordMessage = "The provided password is incorrect."

var errNoAuthenticatedUser = errors.New("no authenticated user, the auth middleware must wrap the handler")

type passwordConfirmer interface {
	Run(ctx context.Context, in auth.ConfirmPasswordIn) (auth.ConfirmPasswordOut, error)
}

type confirmationRecorder interface {
	Confirm(ctx context.Context)
	Intended(ctx context.Context) string
}

type ConfirmPassword struct {
	appName      string
	responder    *web.Responder
	confirm      passwordConfirmer
	confirmation confirmationRecorder
	limit        rateLimit
}

func NewConfirmPassword(
	appName string,
	responder *web.Responder,
	confirm passwordConfirmer,
	confirmation confirmationRecorder,
	limit rateLimit,
) *ConfirmPassword {
	return &ConfirmPassword{
		appName:      appName,
		responder:    responder,
		confirm:      confirm,
		confirmation: confirmation,
		limit:        limit,
	}
}

func (h *ConfirmPassword) Show(w http.ResponseWriter, r *http.Request) error {
	return h.render(w, r, http.StatusOK, confirmpassword.Form{Errors: nil, Alert: ""})
}

func (h *ConfirmPassword) Store(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return fmt.Errorf("parse the password confirmation form: %w", err)
	}

	user, signedIn := middleware.AuthenticatedUser{}.CurrentUser(r.Context())
	if !signedIn {
		return errNoAuthenticatedUser
	}

	err = h.limit.Hit(r.Context(), string(user.ID))
	if err != nil {
		return h.refuse(w, r, err)
	}

	attempt := request.ParseConfirmPassword(r.PostForm)

	invalid := attempt.Validate()
	if len(invalid) > 0 {
		return h.render(w, r, http.StatusUnprocessableEntity, confirmpassword.Form{Errors: invalid, Alert: ""})
	}

	return h.confirmAndContinue(w, r, user, attempt)
}

func (h *ConfirmPassword) confirmAndContinue(
	w http.ResponseWriter,
	r *http.Request,
	user domain.User,
	attempt request.ConfirmPassword,
) error {
	_, err := h.confirm.Run(r.Context(), auth.ConfirmPasswordIn{User: user, Password: attempt.Password})
	if errors.Is(err, auth.ErrIncorrectPassword) {
		incorrect := validate.Errors{request.FieldPassword: incorrectPasswordMessage}

		return h.render(w, r, http.StatusUnprocessableEntity, confirmpassword.Form{Errors: incorrect, Alert: ""})
	}

	if err != nil {
		return fmt.Errorf("confirm the password: %w", err)
	}

	h.confirmation.Confirm(r.Context())
	h.responder.Redirect(w, r, h.confirmation.Intended(r.Context()))

	return nil
}

func (h *ConfirmPassword) refuse(w http.ResponseWriter, r *http.Request, refusal error) error {
	err := h.responder.RateLimited(w, r, refusal, h.page(confirmpassword.Form{Errors: nil, Alert: rateLimitedMessage}))
	if err != nil {
		return fmt.Errorf("refuse the password confirmation: %w", err)
	}

	return nil
}

func (h *ConfirmPassword) render(w http.ResponseWriter, r *http.Request, status int, form confirmpassword.Form) error {
	err := h.responder.Page(w, r, status, h.page(form))
	if err != nil {
		return fmt.Errorf("show the password confirmation page: %w", err)
	}

	return nil
}

func (h *ConfirmPassword) page(form confirmpassword.Form) web.Page {
	return web.Page{Layout: layout.KindPublic, Body: confirmpassword.Page(h.appName, form), Fragment: nil}
}
