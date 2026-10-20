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
	"github.com/DannyFestor/go-template-html/internal/view/register"
	"github.com/DannyFestor/go-template-html/internal/web"
)

const (
	emailTakenMessage  = "This email address is already registered."
	rateLimitedMessage = "Too many attempts. Please try again later."
)

type registrar interface {
	Run(ctx context.Context, in auth.RegisterIn) (auth.RegisterOut, error)
}

type authenticator interface {
	SignIn(ctx context.Context, user domain.User)
}

type rateLimit interface {
	Hit(ctx context.Context, key string) error
}

type Register struct {
	appName       string
	responder     *web.Responder
	register      registrar
	authenticator authenticator
	limit         rateLimit
}

func NewRegister(
	appName string,
	responder *web.Responder,
	register registrar,
	authenticator authenticator,
	limit rateLimit,
) *Register {
	return &Register{
		appName:       appName,
		responder:     responder,
		register:      register,
		authenticator: authenticator,
		limit:         limit,
	}
}

func (h *Register) Show(w http.ResponseWriter, r *http.Request) error {
	return h.render(w, r, http.StatusOK, register.Form{Name: "", Email: "", Errors: nil, Alert: ""})
}

func (h *Register) Store(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return fmt.Errorf("parse the registration form: %w", err)
	}

	registration := request.ParseRegistration(r.PostForm)

	err = h.limit.Hit(r.Context(), middleware.ClientIPFrom(r.Context()).String())
	if err != nil {
		return h.refuse(w, r, err, registration)
	}

	invalid := registration.Validate()
	if len(invalid) > 0 {
		return h.render(w, r, http.StatusUnprocessableEntity, refilled(registration, invalid, ""))
	}

	return h.registerAndSignIn(w, r, registration)
}

func (h *Register) registerAndSignIn(w http.ResponseWriter, r *http.Request, registration request.Registration) error {
	registered, err := h.register.Run(r.Context(), auth.RegisterIn{
		Name:     registration.Name,
		Email:    registration.Email,
		Password: registration.Password,
	})
	if errors.Is(err, domain.ErrEmailTaken) {
		taken := validate.Errors{request.FieldEmail: emailTakenMessage}

		return h.render(w, r, http.StatusUnprocessableEntity, refilled(registration, taken, ""))
	}

	if err != nil {
		return fmt.Errorf("register the visitor: %w", err)
	}

	h.authenticator.SignIn(r.Context(), registered.User)
	h.responder.Redirect(w, r, route.Dashboard)

	return nil
}

func (h *Register) refuse(
	w http.ResponseWriter,
	r *http.Request,
	refusal error,
	registration request.Registration,
) error {
	err := h.responder.RateLimited(w, r, refusal, h.page(refilled(registration, nil, rateLimitedMessage)))
	if err != nil {
		return fmt.Errorf("refuse the registration: %w", err)
	}

	return nil
}

func (h *Register) render(w http.ResponseWriter, r *http.Request, status int, form register.Form) error {
	err := h.responder.Page(w, r, status, h.page(form))
	if err != nil {
		return fmt.Errorf("show the registration page: %w", err)
	}

	return nil
}

func (h *Register) page(form register.Form) web.Page {
	return web.Page{Layout: layout.KindPublic, Body: register.Page(h.appName, form), Fragment: nil}
}

func refilled(registration request.Registration, invalid validate.Errors, alert string) register.Form {
	return register.Form{Name: registration.Name, Email: registration.Email, Errors: invalid, Alert: alert}
}
