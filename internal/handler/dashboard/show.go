package dashboard

import (
	"fmt"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/view/dashboard"
	"github.com/DannyFestor/go-template-html/internal/view/layout"
	"github.com/DannyFestor/go-template-html/internal/web"
)

type Show struct {
	responder *web.Responder
}

func NewShow(responder *web.Responder) *Show {
	return &Show{responder: responder}
}

func (h *Show) Serve(w http.ResponseWriter, r *http.Request) error {
	page := web.Page{Layout: layout.KindApp, Body: dashboard.Page(), Fragment: nil}

	err := h.responder.Page(w, r, http.StatusOK, page)
	if err != nil {
		return fmt.Errorf("show the dashboard: %w", err)
	}

	return nil
}
