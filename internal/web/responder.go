package web

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/view/layout"
)

type errorReporter interface {
	Report(ctx context.Context, err error)
}

type flashSetter interface {
	SetFlash(ctx context.Context, message string)
}

type Responder struct {
	shells           *ShellBuilder
	flashes          flashSetter
	reporter         errorReporter
	showErrorDetails bool
}

func NewResponder(
	shells *ShellBuilder,
	flashes flashSetter,
	reporter errorReporter,
	showErrorDetails bool,
) *Responder {
	return &Responder{shells: shells, flashes: flashes, reporter: reporter, showErrorDetails: showErrorDetails}
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

func (r *Responder) RateLimited(w http.ResponseWriter, req *http.Request, refusal error, page Page) error {
	tooMany, ok := errors.AsType[*StatusError](refusal)
	if !ok || tooMany.Status != http.StatusTooManyRequests {
		return refusal
	}

	setRetryAfter(w.Header(), tooMany.RetryAfter)

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

func (r *Responder) RedirectWithStatus(w http.ResponseWriter, req *http.Request, url, message string) {
	r.flashes.SetFlash(req.Context(), message)
	r.Redirect(w, req, url)
}

func (r *Responder) writeFullPage(
	w http.ResponseWriter,
	req *http.Request,
	status int,
	kind layout.Kind,
	body component,
) error {
	shell, err := r.shells.Build(req.Context())
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
