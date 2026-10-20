package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/DannyFestor/go-template-html/internal/route"
)

type PasswordConfirmation struct {
	clock     clock
	timeout   time.Duration
	responder redirector
	session   Session
	intended  intendedURL
}

func NewPasswordConfirmation(clock clock, timeout time.Duration, responder redirector) *PasswordConfirmation {
	return &PasswordConfirmation{
		clock:     clock,
		timeout:   timeout,
		responder: responder,
		session:   Session{},
		intended:  intendedURL{session: Session{}},
	}
}

func (p *PasswordConfirmation) Confirm(ctx context.Context) {
	p.session.ConfirmPassword(ctx, p.clock.Now())
}

func (p *PasswordConfirmation) Intended(ctx context.Context) string {
	return p.intended.take(ctx)
}

func (p *PasswordConfirmation) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !p.recentlyConfirmed(r.Context()) {
			p.intended.store(r)
			p.responder.Redirect(w, r, route.ConfirmPassword)

			return
		}

		next.ServeHTTP(w, r)
	})
}

func (p *PasswordConfirmation) recentlyConfirmed(ctx context.Context) bool {
	return p.clock.Now().Before(p.session.PasswordConfirmedAt(ctx).Add(p.timeout))
}
