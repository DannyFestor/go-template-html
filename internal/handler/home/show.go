package home

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/view"
	"github.com/DannyFestor/go-template-html/internal/view/home"
	"github.com/DannyFestor/go-template-html/internal/view/layout"
)

type shellBuilder interface {
	Build() (view.Shell, error)
}

type errorReporter interface {
	Report(ctx context.Context, err error)
}

type Show struct {
	shells   shellBuilder
	reporter errorReporter
}

func NewShow(shells shellBuilder, reporter errorReporter) *Show {
	return &Show{shells: shells, reporter: reporter}
}

func (h *Show) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	page, err := h.render(r.Context())
	if err != nil {
		h.reporter.Report(r.Context(), err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = page.WriteTo(w)
}

func (h *Show) render(ctx context.Context) (*bytes.Buffer, error) {
	shell, err := h.shells.Build()
	if err != nil {
		return nil, fmt.Errorf("build the shell: %w", err)
	}

	var page bytes.Buffer

	err = layout.Render(layout.KindPublic, shell, home.Page()).Render(ctx, &page)
	if err != nil {
		return nil, fmt.Errorf("render the home page: %w", err)
	}

	return &page, nil
}
