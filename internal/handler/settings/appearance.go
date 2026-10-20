package settings

import (
	"fmt"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/request"
	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/view/layout"
	"github.com/DannyFestor/go-template-html/internal/view/settings"
	"github.com/DannyFestor/go-template-html/internal/web"
)

type themeCookie interface {
	Read(r *http.Request) domain.Theme
	Issue(w http.ResponseWriter, theme domain.Theme)
}

type Appearance struct {
	responder *web.Responder
	themes    themeCookie
}

func NewAppearance(responder *web.Responder, themes themeCookie) *Appearance {
	return &Appearance{responder: responder, themes: themes}
}

func (h *Appearance) Show(w http.ResponseWriter, r *http.Request) error {
	return h.render(w, r, http.StatusOK, settings.AppearanceForm{Theme: string(h.themes.Read(r)), Errors: nil})
}

func (h *Appearance) Update(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return fmt.Errorf("parse the appearance form: %w", err)
	}

	appearance := request.ParseAppearance(r.PostForm)

	invalid := appearance.Validate()
	if len(invalid) > 0 {
		form := settings.AppearanceForm{Theme: string(h.themes.Read(r)), Errors: invalid}

		return h.render(w, r, http.StatusUnprocessableEntity, form)
	}

	h.themes.Issue(w, appearance.Theme)
	h.responder.Redirect(w, r, route.SettingsAppearance)

	return nil
}

func (h *Appearance) render(w http.ResponseWriter, r *http.Request, status int, form settings.AppearanceForm) error {
	page := web.Page{Layout: layout.KindApp, Body: settings.Appearance(form), Fragment: nil}

	err := h.responder.Page(w, r, status, page)
	if err != nil {
		return fmt.Errorf("show the appearance settings: %w", err)
	}

	return nil
}
