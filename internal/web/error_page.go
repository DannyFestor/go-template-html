package web

import (
	"net/http"
	"runtime/debug"

	"github.com/DannyFestor/go-template-html/internal/view/errorpage"
	"github.com/DannyFestor/go-template-html/internal/view/layout"
)

func (r *Responder) Error(w http.ResponseWriter, req *http.Request, err error) {
	rendered, expected := classify(err)
	if !expected {
		r.reporter.Report(req.Context(), err)
	}

	setErrorHeaders(w.Header(), req, rendered)
	r.writeErrorPage(w, req, errorModel(rendered, r.errorDetail(expected, err)))
}

func (r *Responder) writeErrorPage(w http.ResponseWriter, req *http.Request, model errorpage.Model) {
	err := r.writeFullPage(w, req, model.Status, layout.KindPublic, errorpage.Page(model))
	if err != nil {
		r.reporter.Report(req.Context(), err)
		http.Error(w, model.Title, model.Status)
	}
}

func (r *Responder) errorDetail(expected bool, err error) string {
	if expected || !r.showErrorDetails {
		return ""
	}

	return err.Error() + "\n\n" + string(debug.Stack())
}

func errorModel(rendered *StatusError, detail string) errorpage.Model {
	return errorpage.Model{
		Status:  rendered.Status,
		Title:   http.StatusText(rendered.Status),
		Message: errorMessage(rendered.Status),
		Detail:  detail,
	}
}

func setErrorHeaders(header http.Header, req *http.Request, rendered *StatusError) {
	if isHTMX(req) {
		swapWholeBody(header)
	}

	if rendered.RetryAfter > 0 {
		setRetryAfter(header, rendered.RetryAfter)
	}
}

func errorMessage(status int) string {
	switch status {
	case http.StatusForbidden:
		return "You are not allowed to do this."
	case http.StatusNotFound:
		return "This page does not exist."
	case http.StatusMethodNotAllowed:
		return "This page does not accept this kind of request."
	case http.StatusRequestEntityTooLarge:
		return "The submitted data is too large."
	case http.StatusTooManyRequests:
		return "Too many attempts. Please try again later."
	case http.StatusServiceUnavailable:
		return "The service is unavailable. Please try again shortly."
	default:
		return "Something went wrong on our side."
	}
}
