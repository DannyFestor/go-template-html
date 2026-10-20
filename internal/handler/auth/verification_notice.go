package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/view/layout"
	"github.com/DannyFestor/go-template-html/internal/view/verifyemail"
	"github.com/DannyFestor/go-template-html/internal/web"
)

const verificationLinkSentMessage = "A new verification link has been sent to the email address you provided " +
	"during registration."

type verificationSender interface {
	Run(ctx context.Context, in auth.SendVerificationMailIn) (auth.SendVerificationMailOut, error)
}

type VerificationNotice struct {
	appName   string
	responder *web.Responder
	send      verificationSender
	limit     rateLimit
}

func NewVerificationNotice(
	appName string,
	responder *web.Responder,
	send verificationSender,
	limit rateLimit,
) *VerificationNotice {
	return &VerificationNotice{appName: appName, responder: responder, send: send, limit: limit}
}

func (h *VerificationNotice) Show(w http.ResponseWriter, r *http.Request) error {
	if h.sentAwayVerified(w, r) {
		return nil
	}

	return h.render(w, r, http.StatusOK, verifyemail.Notice{Alert: ""})
}

func (h *VerificationNotice) Store(w http.ResponseWriter, r *http.Request) error {
	if h.sentAwayVerified(w, r) {
		return nil
	}

	return h.resend(w, r, authenticatedUser(r))
}

func (h *VerificationNotice) sentAwayVerified(w http.ResponseWriter, r *http.Request) bool {
	verified := authenticatedUser(r).Verified()
	if verified {
		h.responder.Redirect(w, r, route.Dashboard)
	}

	return verified
}

func (h *VerificationNotice) resend(w http.ResponseWriter, r *http.Request, user domain.User) error {
	err := h.limit.Hit(r.Context(), string(user.ID))
	if err != nil {
		return h.refuse(w, r, err)
	}

	_, err = h.send.Run(r.Context(), auth.SendVerificationMailIn{User: user})
	if err != nil {
		return fmt.Errorf("resend the verification mail: %w", err)
	}

	h.responder.RedirectWithStatus(w, r, route.VerifyEmail, verificationLinkSentMessage)

	return nil
}

func (h *VerificationNotice) refuse(w http.ResponseWriter, r *http.Request, refusal error) error {
	err := h.responder.RateLimited(w, r, refusal, h.page(verifyemail.Notice{Alert: rateLimitedMessage}))
	if err != nil {
		return fmt.Errorf("refuse the verification mail: %w", err)
	}

	return nil
}

func (h *VerificationNotice) render(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	notice verifyemail.Notice,
) error {
	err := h.responder.Page(w, r, status, h.page(notice))
	if err != nil {
		return fmt.Errorf("show the verification notice: %w", err)
	}

	return nil
}

func (h *VerificationNotice) page(notice verifyemail.Notice) web.Page {
	return web.Page{Layout: layout.KindPublic, Body: verifyemail.Page(h.appName, notice), Fragment: nil}
}
