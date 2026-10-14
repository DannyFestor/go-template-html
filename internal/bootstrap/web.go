package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/adapter/logreport"
	"github.com/DannyFestor/go-template-html/internal/buildinfo"
	"github.com/DannyFestor/go-template-html/internal/handler/health"
	"github.com/DannyFestor/go-template-html/internal/handler/home"
	"github.com/DannyFestor/go-template-html/internal/web"
)

func RunWeb() int {
	return run(serveWeb)
}

func serveWeb(ctx context.Context, config Config, logger *slog.Logger) error {
	handler, err := NewWebHandler(config, logger)
	if err != nil {
		return err
	}

	var listenConfig net.ListenConfig

	listener, err := listenConfig.Listen(ctx, "tcp", config.HTTP.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", config.HTTP.Addr, err)
	}

	logger.InfoContext(ctx, "web started",
		slog.String("version", buildinfo.Version),
		slog.String("addr", listener.Addr().String()),
	)

	server := newWebServer(config.HTTP, handler, logger)

	return Serve(ctx, server, listener, config.App.ShutdownTimeout)
}

func NewWebHandler(config Config, logger *slog.Logger) (*http.ServeMux, error) {
	assets, err := newAssetServer(config.App.Env)
	if err != nil {
		return nil, err
	}

	shells := web.NewShellBuilder(config.App.Name, assets)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health.Liveness)
	mux.Handle("GET /assets/", assets)
	mux.Handle("GET /{$}", home.NewShow(shells, logreport.New(logger)))

	return mux, nil
}

func newWebServer(config HTTP, handler http.Handler, logger *slog.Logger) *http.Server {
	return &http.Server{
		Addr:                         "",
		Handler:                      handler,
		DisableGeneralOptionsHandler: false,
		TLSConfig:                    nil,
		ReadTimeout:                  0,
		ReadHeaderTimeout:            config.ReadHeaderTimeout,
		WriteTimeout:                 0,
		IdleTimeout:                  0,
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
