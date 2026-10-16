//go:build feature

package feature_test

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/bootstrap"
	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/web"
)

const (
	probeFailure  = "/probe/failure"
	probeRedirect = "/probe/redirect"
	probeForm     = "/probe/form"
)

var errProbeFailed = errors.New("probe failed")

func mountProbes(webApp *bootstrap.WebApp) {
	responder := webApp.Responder

	webApp.Mux.Handle("GET "+probeFailure, responder.Handle(failing))
	webApp.Mux.Handle("POST "+probeRedirect, responder.Handle(redirectingHome(responder)))
	webApp.Mux.Handle("POST "+probeForm, responder.Handle(parsingTheForm))
}

func failing(http.ResponseWriter, *http.Request) error {
	return errProbeFailed
}

func redirectingHome(responder *web.Responder) web.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		responder.Redirect(w, r, route.Home)

		return nil
	}
}

func parsingTheForm(w http.ResponseWriter, r *http.Request) error {
	err := r.ParseForm()
	if err != nil {
		return fmt.Errorf("parse the form: %w", err)
	}

	w.WriteHeader(http.StatusNoContent)

	return nil
}
