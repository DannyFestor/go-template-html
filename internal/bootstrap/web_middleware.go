package bootstrap

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"

	"github.com/DannyFestor/go-template-html/internal/adapter/clock"
	"github.com/DannyFestor/go-template-html/internal/adapter/compression"
	"github.com/DannyFestor/go-template-html/internal/adapter/valkey"
	"github.com/DannyFestor/go-template-html/internal/middleware"
	"github.com/DannyFestor/go-template-html/internal/web"
)

type wrapper func(http.Handler) http.Handler

type webMiddlewareDeps struct {
	config    Config
	logger    *slog.Logger
	responder *web.Responder
	sessions  *valkey.SessionStore
}

// webMiddleware lists the chain outermost first. Compression wraps everything, so the pages recovery renders are
// compressed too. Recovery and the access log wrap everything that can fail, so a 500 is logged, and the mux error
// interceptor sits right around the mux, so later middleware can enrich its pages.
func webMiddleware(deps webMiddlewareDeps) ([]wrapper, error) {
	compress, err := compression.New()
	if err != nil {
		return nil, fmt.Errorf("compress responses: %w", err)
	}

	crossOrigin, err := middleware.CrossOriginProtection(deps.config.HTTP.CSRFTrustedOrigins, deps.responder)
	if err != nil {
		return nil, fmt.Errorf("protect against cross-origin requests: %w", err)
	}

	return []wrapper{
		compress,
		middleware.RequestID,
		middleware.AccessLog(deps.logger),
		middleware.Recover(deps.responder),
		middleware.SecurityHeaders(deps.config.App.servedOverHTTPS()),
		middleware.LimitBody(deps.config.HTTP.MaxBodyBytes),
		crossOrigin,
		middleware.ClientIP(deps.config.HTTP.TrustedProxies),
		middleware.LoadAndSaveSession(deps.sessions, clock.System{}, sessionSettings(deps.config), deps.responder),
		middleware.InterceptMuxErrors(deps.responder),
	}, nil
}

func sessionSettings(config Config) middleware.SessionSettings {
	return middleware.SessionSettings{
		Secure:           config.App.servedOverHTTPS(),
		AbsoluteLifetime: config.Session.AbsoluteLifetime,
	}
}

func chain(wrappers ...wrapper) wrapper {
	return func(handler http.Handler) http.Handler {
		return withMiddleware(handler, wrappers)
	}
}

func withMiddleware(handler http.Handler, chain []wrapper) http.Handler {
	for _, wrap := range slices.Backward(chain) {
		handler = wrap(handler)
	}

	return handler
}
