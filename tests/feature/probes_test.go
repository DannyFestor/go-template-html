//go:build feature

package feature_test

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/DannyFestor/go-template-html/internal/bootstrap"
	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/web"
)

const (
	probeFailure       = "/probe/failure"
	probeRedirect      = "/probe/redirect"
	probeForm          = "/probe/form"
	probeFlash         = "/probe/flash"
	probeLargePage     = "/probe/large-page"
	probeSecretPage    = "/probe/secret-page"
	probeSensitivePage = "/probe/sensitive-page"

	probeFlashMessage    = "Probe flashed."
	sensitivePageContent = "Sensitive."
)

var errProbeFailed = errors.New("probe failed")

func mountProbes(webApp *bootstrap.WebApp) {
	responder := webApp.Responder

	webApp.Mux.Handle("GET "+probeFailure, responder.Handle(failing))
	webApp.Mux.Handle("POST "+probeRedirect, responder.Handle(redirectingHome(responder)))
	webApp.Mux.Handle("POST "+probeForm, responder.Handle(parsingTheForm))
	webApp.Mux.Handle("POST "+probeFlash, responder.Handle(flashingHome(responder)))
	webApp.Mux.HandleFunc("GET "+probeLargePage, writingTheLargePage)
	webApp.Mux.HandleFunc("GET "+probeSecretPage, writingTheSecretPage)
	webApp.Mux.Handle("GET "+probeSensitivePage,
		webApp.RequirePasswordConfirmation(http.HandlerFunc(writingTheSensitivePage)))
}

func writingTheSensitivePage(w http.ResponseWriter, _ *http.Request) {
	_, _ = io.WriteString(w, sensitivePageContent)
}

func flashingHome(responder *web.Responder) web.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		responder.RedirectWithStatus(w, r, route.Home, probeFlashMessage)

		return nil
	}
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

func writingTheSecretPage(w http.ResponseWriter, r *http.Request) {
	web.SkipCompression(w)
	writingTheLargePage(w, r)
}

func writingTheLargePage(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, largePage())
}

func largePage() string {
	return "<!doctype html><title>Large</title><p>" + strings.Repeat("Compressible text. ", 200) + "</p>"
}
