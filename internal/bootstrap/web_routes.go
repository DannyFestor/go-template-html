package bootstrap

import (
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/asset"
	"github.com/DannyFestor/go-template-html/internal/handler/health"
	"github.com/DannyFestor/go-template-html/internal/handler/home"
	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/web"
)

type webRoutes struct {
	assets    *asset.Server
	responder *web.Responder
	readiness *health.Readiness
}

func newWebMux(routes webRoutes) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+route.Healthz, health.Liveness)
	mux.Handle("GET "+route.Readyz, routes.readiness)
	mux.Handle("GET "+route.Assets, routes.assets)
	mux.Handle("GET "+route.Home+"{$}", routes.responder.Handle(home.NewShow(routes.responder).Serve))

	return mux
}
