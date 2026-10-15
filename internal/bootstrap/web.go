package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/adapter/logreport"
	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/buildinfo"
	"github.com/DannyFestor/go-template-html/internal/handler/health"
	"github.com/DannyFestor/go-template-html/internal/web"
)

type WebApp struct {
	Mux       *http.ServeMux
	Responder *web.Responder
	handler   http.Handler
	pool      *postgres.Pool
}

func RunWeb() int {
	return run(serveWeb)
}

func serveWeb(ctx context.Context, config Config, logger *slog.Logger) error {
	app, err := NewWebApp(ctx, config, logger)
	if err != nil {
		return err
	}
	defer app.Close()

	var listenConfig net.ListenConfig

	listener, err := listenConfig.Listen(ctx, "tcp", config.HTTP.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", config.HTTP.Addr, err)
	}

	logger.InfoContext(ctx, "web started",
		slog.String("version", buildinfo.Version),
		slog.String("addr", listener.Addr().String()),
	)

	server := newWebServer(config.HTTP, app.Handler(), logger)

	return Serve(ctx, server, listener, config.App.ShutdownTimeout)
}

func NewWebApp(ctx context.Context, config Config, logger *slog.Logger) (*WebApp, error) {
	assets, err := newAssetServer(config.App.Env)
	if err != nil {
		return nil, err
	}

	pool, err := postgres.Open(ctx, webPostgresConfig(config.DB))
	if err != nil {
		return nil, fmt.Errorf("connect to the database: %w", err)
	}

	responder := web.NewResponder(
		web.NewShellBuilder(config.App.Name, assets),
		logreport.New(logger),
		config.App.Env == EnvironmentLocal,
	)
	mux := newWebMux(webRoutes{
		assets:    assets,
		responder: responder,
		readiness: health.NewReadiness(health.NewCachedPing(pool, config.HTTP.ReadinessCacheTTL)),
	})

	return &WebApp{
		Mux:       mux,
		Responder: responder,
		handler:   withMiddleware(mux, webMiddleware(config.HTTP, logger, responder)),
		pool:      pool,
	}, nil
}

func (a *WebApp) Handler() http.Handler {
	return a.handler
}

func (a *WebApp) Close() {
	a.pool.Close()
}

func newWebServer(config HTTP, handler http.Handler, logger *slog.Logger) *http.Server {
	return &http.Server{
		Addr:                         "",
		Handler:                      handler,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadTimeout:                  config.ReadTimeout,
		ReadHeaderTimeout:            config.ReadHeaderTimeout,
		WriteTimeout:                 config.WriteTimeout,
		IdleTimeout:                  config.IdleTimeout,
		MaxHeaderBytes:               0,
		MaxHeaderValueCount:          0,
		TLSNextProto:                 nil,
		ConnState:                    nil,
		ErrorLog:                     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		BaseContext:                  nil,
		ConnContext:                  nil,
		HTTP2:                        nil,
		Protocols:                    nil,
		DisableClientPriority:        false,
	}
}
