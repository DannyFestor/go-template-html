package web

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/DannyFestor/go-template-html/internal/view/layout"
)

type errorReporter interface {
	Report(ctx context.Context, err error)
}

type Responder struct {
	shells           *ShellBuilder
	reporter         errorReporter
	showErrorDetails bool
}

func NewResponder(shells *ShellBuilder, reporter errorReporter, showErrorDetails bool) *Responder {
	return &Responder{shells: shells, reporter: reporter, showErrorDetails: showErrorDetails}
}

func (r *Responder) Page(w http.ResponseWriter, req *http.Request, status int, page Page) error {
	if isPartial(req) && page.Fragment != nil {
		return writeHTML(req.Context(), w, status, page.Fragment)
	}

	if crossesLayouts(req, page.Layout) {
		swapWholeBody(w.Header())
	}

	return r.writeFullPage(w, req, status, page.Layout, page.Body)
}

func (r *Responder) TooManyRequests(
	w http.ResponseWriter,
	req *http.Request,
	retryAfter time.Duration,
	page Page,
) error {
	setRetryAfter(w.Header(), retryAfter)

	return r.Page(w, req, http.StatusTooManyRequests, page)
}

// Redirect answers partial requests with HX-Redirect, because htmx would swap a followed 303 into the target and
// ignores HX-Redirect on a 3xx.
func (r *Responder) Redirect(w http.ResponseWriter, req *http.Request, url string) {
	if isPartial(req) {
		w.Header().Set(headerRedirect, url)
		w.WriteHeader(http.StatusOK)

		return
	}

	http.Redirect(w, req, url, http.StatusSeeOther)
}

func (r *Responder) writeFullPage(
	w http.ResponseWriter,
	req *http.Request,
	status int,
	kind layout.Kind,
	body component,
) error {
	shell, err := r.shells.Build()
	if err != nil {
		return fmt.Errorf("build the shell: %w", err)
	}

	return writeHTML(req.Context(), w, status, layout.Render(kind, shell, body))
}

// writeHTML renders into a buffer first, so a failing render can still become an error page.
func writeHTML(ctx context.Context, w http.ResponseWriter, status int, content component) error {
	var buffer bytes.Buffer

	err := content.Render(ctx, &buffer)
	if err != nil {
		return fmt.Errorf("render the page: %w", err)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buffer.WriteTo(w)

	return nil
}
