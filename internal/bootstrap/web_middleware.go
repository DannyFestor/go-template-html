package bootstrap

import (
	"log/slog"
	"net/http"
	"slices"

	"github.com/DannyFestor/go-template-html/internal/middleware"
	"github.com/DannyFestor/go-template-html/internal/web"
)

type wrapper func(http.Handler) http.Handler

// webMiddleware lists the chain outermost first. Recovery and the access log wrap everything that can fail, so a 500
// is logged, and the mux error interceptor sits right around the mux, so later middleware can enrich its pages.
func webMiddleware(config HTTP, logger *slog.Logger, responder *web.Responder) []wrapper {
	return []wrapper{
		middleware.RequestID,
		middleware.AccessLog(logger),
		middleware.Recover(responder),
		middleware.LimitBody(config.MaxBodyBytes),
		middleware.InterceptMuxErrors(responder),
	}
}

func withMiddleware(handler http.Handler, chain []wrapper) http.Handler {
	for _, wrap := range slices.Backward(chain) {
		handler = wrap(handler)
	}

	return handler
}
