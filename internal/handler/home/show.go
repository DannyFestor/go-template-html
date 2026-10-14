package home

import (
	"fmt"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/view/home"
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
	err := h.responder.Page(w, r, http.StatusOK, web.Page{Layout: layout.KindPublic, Body: home.Page(), Fragment: nil})
	if err != nil {
		return fmt.Errorf("show the home page: %w", err)
	}

	return nil
}
