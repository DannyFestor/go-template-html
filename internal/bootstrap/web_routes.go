package bootstrap

import (
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/asset"
	"github.com/DannyFestor/go-template-html/internal/handler/health"
	"github.com/DannyFestor/go-template-html/internal/handler/home"
	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/web"
)

func newWebMux(assets *asset.Server, responder *web.Responder) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET "+route.Healthz, health.Liveness)
	mux.Handle("GET "+route.Assets, assets)
	mux.Handle("GET "+route.Home+"{$}", responder.Handle(home.NewShow(responder).Serve))

	return mux
}
