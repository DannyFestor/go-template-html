package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/buildinfo"
	"github.com/DannyFestor/go-template-html/internal/handler/health"
)

func RunWeb() int {
	return run(serveWeb)
}

func serveWeb(ctx context.Context, config Config, logger *slog.Logger) error {
	var listenConfig net.ListenConfig

	listener, err := listenConfig.Listen(ctx, "tcp", config.HTTP.Addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", config.HTTP.Addr, err)
	}

	logger.InfoContext(ctx, "web started",
		slog.String("version", buildinfo.Version),
		slog.String("addr", listener.Addr().String()),
	)

	server := newWebServer(config.HTTP, webRoutes(), logger)

	return Serve(ctx, server, listener, config.App.ShutdownTimeout)
}

func webRoutes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health.Liveness)

	return mux
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
