package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/web"
)

type recaller interface {
	Run(ctx context.Context, in auth.RecallIn) (auth.RecallOut, error)
}

type logoutAction interface {
	Run(ctx context.Context, in auth.LogoutIn) (auth.LogoutOut, error)
}

type Sessions struct {
	responder     *web.Responder
	recall        recaller
	logout        logoutAction
	authenticator authenticator
	cookie        rememberCookie
}

func NewSessions(
	responder *web.Responder,
	recall recaller,
	logout logoutAction,
	authenticator authenticator,
	cookie rememberCookie,
) *Sessions {
	return &Sessions{
		responder:     responder,
		recall:        recall,
		logout:        logout,
		authenticator: authenticator,
		cookie:        cookie,
	}
}

func (s *Sessions) SignIn(w http.ResponseWriter, r *http.Request, user domain.User, token domain.RememberToken) {
	s.authenticator.SignIn(r.Context(), user)

	var notRemembered domain.RememberToken
	if token != notRemembered {
		s.cookie.Issue(w, token)
	}
}

func (s *Sessions) SignOut(w http.ResponseWriter, r *http.Request) error {
	token, remembered := s.cookie.Read(r)
	if remembered {
		_, err := s.logout.Run(r.Context(), auth.LogoutIn{RememberToken: token})
		if err != nil {
			return fmt.Errorf("forget the remembered device: %w", err)
		}

		s.cookie.Clear(w)
	}

	s.authenticator.SignOut(r.Context())

	return nil
}

func (s *Sessions) Intended(ctx context.Context) string {
	return s.authenticator.Intended(ctx)
}

func (s *Sessions) RestoreRemembered(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := s.signInRemembered(w, r)
		if err != nil {
			s.responder.Error(w, r, err)

			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Sessions) signInRemembered(w http.ResponseWriter, r *http.Request) error {
	if s.authenticator.SignedIn(r.Context()) {
		return nil
	}

	token, remembered := s.cookie.Read(r)
	if !remembered {
		return nil
	}

	recalled, err := s.recall.Run(r.Context(), auth.RecallIn{Token: token})
	if errors.Is(err, auth.ErrNotRemembered) {
		s.cookie.Clear(w)

		return nil
	}

	if err != nil {
		return fmt.Errorf("restore the remembered session: %w", err)
	}

	s.SignIn(w, r, recalled.User, recalled.Token)

	return nil
}
