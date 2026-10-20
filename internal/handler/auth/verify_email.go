package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/web"
)

const emailVerifiedMessage = "Your email address has been verified."

type emailVerification interface {
	Run(ctx context.Context, in auth.VerifyEmailIn) (auth.VerifyEmailOut, error)
}

type intendedDestination interface {
	Intended(ctx context.Context) string
}

type VerifyEmail struct {
	responder   *web.Responder
	verify      emailVerification
	destination intendedDestination
	limit       rateLimit
}

func NewVerifyEmail(
	responder *web.Responder,
	verify emailVerification,
	destination intendedDestination,
	limit rateLimit,
) *VerifyEmail {
	return &VerifyEmail{responder: responder, verify: verify, destination: destination, limit: limit}
}

func (h *VerifyEmail) Serve(w http.ResponseWriter, r *http.Request) error {
	user := authenticatedUser(r)

	err := h.limit.Hit(r.Context(), string(user.ID))
	if err != nil {
		return fmt.Errorf("refuse the email verification: %w", err)
	}

	_, err = h.verify.Run(r.Context(), auth.VerifyEmailIn{
		User:       user,
		LinkUserID: domain.UserID(r.PathValue(route.VerifyEmailUserID)),
		LinkHash:   r.PathValue(route.VerifyEmailHash),
	})
	if errors.Is(err, auth.ErrVerificationLinkMismatch) {
		return web.ErrForbidden
	}

	if err != nil {
		return fmt.Errorf("verify the email address: %w", err)
	}

	h.responder.RedirectWithStatus(w, r, h.destination.Intended(r.Context()), emailVerifiedMessage)

	return nil
}
