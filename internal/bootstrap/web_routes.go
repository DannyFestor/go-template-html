package bootstrap

import (
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/asset"
	dashboardhandler "github.com/DannyFestor/go-template-html/internal/handler/dashboard"
	healthhandler "github.com/DannyFestor/go-template-html/internal/handler/health"
	homehandler "github.com/DannyFestor/go-template-html/internal/handler/home"
	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/web"
)

type webRoutes struct {
	assets    *asset.Server
	responder *web.Responder
	readiness *healthhandler.Readiness
	auth      authRoutes
}

func newWebMux(routes webRoutes) *http.ServeMux {
	responder := routes.responder

	mux := http.NewServeMux()
	mux.HandleFunc("GET "+route.Healthz, healthhandler.Liveness)
	mux.Handle("GET "+route.Readyz, routes.readiness)
	mux.Handle("GET "+route.Assets, routes.assets)
	mux.Handle("GET "+route.Home+"{$}", responder.Handle(homehandler.NewShow(responder).Serve))
	mux.Handle("GET "+route.Register, routes.auth.requireGuest(responder.Handle(routes.auth.register.Show)))
	mux.Handle("POST "+route.Register, routes.auth.requireGuest(responder.Handle(routes.auth.register.Store)))
	mux.Handle("GET "+route.Login, routes.auth.requireGuest(responder.Handle(routes.auth.login.Show)))
	mux.Handle("POST "+route.Login, routes.auth.requireGuest(responder.Handle(routes.auth.login.Store)))
	mux.Handle("POST "+route.Logout, routes.auth.requireUser(responder.Handle(routes.auth.logout.Serve)))
	mux.Handle("GET "+route.ConfirmPassword,
		routes.auth.requireUser(responder.Handle(routes.auth.confirmPassword.Show)))
	mux.Handle("POST "+route.ConfirmPassword,
		routes.auth.requireUser(responder.Handle(routes.auth.confirmPassword.Store)))
	mux.Handle("GET "+route.Dashboard,
		routes.auth.requireUser(responder.Handle(dashboardhandler.NewShow(responder).Serve)))

	return mux
}
