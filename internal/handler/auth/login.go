package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/middleware"
	"github.com/DannyFestor/go-template-html/internal/request"
	"github.com/DannyFestor/go-template-html/internal/view/layout"
	"github.com/DannyFestor/go-template-html/internal/view/login"
	"github.com/DannyFestor/go-template-html/internal/web"
)

const noMatchingAccountMessage = "These credentials do not match our records."

type loginAction interface {
	Run(ctx context.Context, in auth.LoginIn) (auth.LoginOut, error)
}

type sessionStarter interface {
	SignIn(w http.ResponseWriter, r *http.Request, user domain.User, token domain.RememberToken)
	Intended(ctx context.Context) string
}

type clearableRateLimit interface {
	rateLimit

	Clear(ctx context.Context, key string) error
}

type Login struct {
	appName   string
	responder *web.Responder
	login     loginAction
	sessions  sessionStarter
	limit     clearableRateLimit
}

func NewLogin(
	appName string,
	responder *web.Responder,
	login loginAction,
	sessions sessionStarter,
	limit clearableRateLimit,
) *Login {
	return &Login{appName: appName, responder: responder, login: login, sessions: sessions, limit: limit}
}

func (h *Login) Show(w http.ResponseWriter, r *http.Request) error {
	return h.render(w, r, http.StatusOK, login.Form{Email: "", Remember: false, Errors: nil, Alert: ""})
}

func (h *Login) Store(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return fmt.Errorf("parse the login form: %w", err)
	}

	attempt := request.ParseLogin(r.PostForm)

	err = h.limit.Hit(r.Context(), loginLimitKey(r.Context(), attempt))
	if err != nil {
		return h.refuse(w, r, err, attempt)
	}

	invalid := attempt.Validate()
	if len(invalid) > 0 {
		return h.render(w, r, http.StatusUnprocessableEntity, refilledLogin(attempt, invalid, ""))
	}

	return h.logIn(w, r, attempt)
}

func (h *Login) logIn(w http.ResponseWriter, r *http.Request, attempt request.Login) error {
	loggedIn, err := h.login.Run(r.Context(), auth.LoginIn{
		Email:    attempt.Email,
		Password: attempt.Password,
		Remember: attempt.Remember,
	})
	if errors.Is(err, auth.ErrInvalidCredentials) {
		return h.render(w, r, http.StatusUnprocessableEntity, refilledLogin(attempt, nil, noMatchingAccountMessage))
	}

	if err != nil {
		return fmt.Errorf("log the user in: %w", err)
	}

	err = h.limit.Clear(r.Context(), loginLimitKey(r.Context(), attempt))
	if err != nil {
		return fmt.Errorf("clear the login attempts: %w", err)
	}

	h.sessions.SignIn(w, r, loggedIn.User, loggedIn.RememberToken)
	h.responder.Redirect(w, r, h.sessions.Intended(r.Context()))

	return nil
}

func (h *Login) refuse(w http.ResponseWriter, r *http.Request, refusal error, attempt request.Login) error {
	err := h.responder.RateLimited(w, r, refusal, h.page(refilledLogin(attempt, nil, rateLimitedMessage)))
	if err != nil {
		return fmt.Errorf("refuse the login: %w", err)
	}

	return nil
}

func (h *Login) render(w http.ResponseWriter, r *http.Request, status int, form login.Form) error {
	err := h.responder.Page(w, r, status, h.page(form))
	if err != nil {
		return fmt.Errorf("show the login page: %w", err)
	}

	return nil
}

func (h *Login) page(form login.Form) web.Page {
	return web.Page{Layout: layout.KindPublic, Body: login.Page(h.appName, form), Fragment: nil}
}

func loginLimitKey(ctx context.Context, attempt request.Login) string {
	return strings.ToLower(attempt.Email) + "|" + middleware.ClientIPFrom(ctx).String()
}

func refilledLogin(attempt request.Login, invalid map[string]string, alert string) login.Form {
	return login.Form{Email: attempt.Email, Remember: attempt.Remember, Errors: invalid, Alert: alert}
}
