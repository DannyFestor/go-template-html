package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/route"
)

type userFinder interface {
	Find(ctx context.Context, id domain.UserID) (domain.User, bool, error)
}

type credentialFingerprints interface {
	Of(hash domain.PasswordHash) string
	Matches(hash domain.PasswordHash, fingerprint string) bool
}

type redirector interface {
	Redirect(w http.ResponseWriter, r *http.Request, url string)
}

type authResponder interface {
	errorRenderer
	redirector
}

type Authenticator struct {
	users       userFinder
	credentials credentialFingerprints
	responder   authResponder
	session     Session
	intended    intendedURL
}

func NewAuthenticator(users userFinder, credentials credentialFingerprints, responder authResponder) *Authenticator {
	return &Authenticator{
		users:       users,
		credentials: credentials,
		responder:   responder,
		session:     Session{},
		intended:    intendedURL{session: Session{}},
	}
}

func (a *Authenticator) SignIn(ctx context.Context, user domain.User) {
	a.session.SignIn(ctx, string(user.ID), a.credentials.Of(user.PasswordHash))
}

func (a *Authenticator) SignOut(ctx context.Context) {
	a.session.SignOut(ctx)
}

func (a *Authenticator) SignedIn(ctx context.Context) bool {
	_, signedIn := a.session.CurrentUserID(ctx)

	return signedIn
}

func (a *Authenticator) Intended(ctx context.Context) string {
	return a.intended.take(ctx)
}

func (a *Authenticator) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, authenticated, err := a.authenticate(r.Context())
		if err != nil {
			a.responder.Error(w, r, err)

			return
		}

		if !authenticated {
			a.sendAway(w, r, route.Login)

			return
		}

		next.ServeHTTP(w, r.WithContext(withAuthenticatedUser(r.Context(), user)))
	})
}

func (a *Authenticator) RequireGuest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, authenticated, err := a.authenticate(r.Context())
		if err != nil {
			a.responder.Error(w, r, err)

			return
		}

		if authenticated {
			a.responder.Redirect(w, r, route.Dashboard)

			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *Authenticator) RequireVerified(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, _ := AuthenticatedUser{}.CurrentUser(r.Context())
		if !user.Verified() {
			a.sendAway(w, r, route.VerifyEmail)

			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *Authenticator) sendAway(w http.ResponseWriter, r *http.Request, target string) {
	a.intended.store(r)
	a.responder.Redirect(w, r, target)
}

func (a *Authenticator) authenticate(ctx context.Context) (domain.User, bool, error) {
	var guest domain.User

	userID, signedIn := a.session.CurrentUserID(ctx)
	if !signedIn {
		return guest, false, nil
	}

	user, found, err := a.users.Find(ctx, domain.UserID(userID))
	if err != nil {
		return guest, false, fmt.Errorf("authenticate the signed-in user: %w", err)
	}

	if !found || !a.credentials.Matches(user.PasswordHash, a.session.CredentialFingerprint(ctx)) {
		a.session.SignOut(ctx)

		return guest, false, nil
	}

	return user, true, nil
}
