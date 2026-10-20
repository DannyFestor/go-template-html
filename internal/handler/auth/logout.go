package auth

import (
	"fmt"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/web"
)

type sessionEnder interface {
	SignOut(w http.ResponseWriter, r *http.Request) error
}

type Logout struct {
	responder *web.Responder
	sessions  sessionEnder
}

func NewLogout(responder *web.Responder, sessions sessionEnder) *Logout {
	return &Logout{responder: responder, sessions: sessions}
}

func (h *Logout) Serve(w http.ResponseWriter, r *http.Request) error {
	err := h.sessions.SignOut(w, r)
	if err != nil {
		return fmt.Errorf("log the user out: %w", err)
	}

	h.responder.Redirect(w, r, route.Home)

	return nil
}
