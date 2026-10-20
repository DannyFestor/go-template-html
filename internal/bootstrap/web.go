package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/adapter/argon2id"
	"github.com/DannyFestor/go-template-html/internal/adapter/logreport"
	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/adapter/valkey"
	"github.com/DannyFestor/go-template-html/internal/asset"
	"github.com/DannyFestor/go-template-html/internal/buildinfo"
	healthhandler "github.com/DannyFestor/go-template-html/internal/handler/health"
	settingshandler "github.com/DannyFestor/go-template-html/internal/handler/settings"
	"github.com/DannyFestor/go-template-html/internal/middleware"
	"github.com/DannyFestor/go-template-html/internal/web"
)

type WebApp struct {
	Mux                         *http.ServeMux
	Jobs                        *postgres.JobQueue
	Users                       *postgres.UserStore
	RememberTokens              *postgres.RememberTokenStore
	Passwords                   *argon2id.Hasher
	Responder                   *web.Responder
	RequirePasswordConfirmation func(http.Handler) http.Handler
	handler                     http.Handler
	pool                        *postgres.Pool
	valkey                      *valkey.Client
}

func RunWeb() int {
	return run(serveWeb)
}

func serveWeb(ctx context.Context, config Config, logger *slog.Logger) error {
	app, err := NewWebApp(ctx, config, logger)
	if err != nil {
		return err
	}

	return errors.Join(serveWebApp(ctx, config, logger, app), app.Close())
}

func serveWebApp(ctx context.Context, config Config, logger *slog.Logger, app *WebApp) error {
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

	passwords, err := newPasswordHasher(config.PasswordHashing)
	if err != nil {
		return nil, err
	}

	app := &WebApp{
		Mux:                         nil,
		Jobs:                        nil,
		Users:                       nil,
		RememberTokens:              nil,
		Passwords:                   passwords,
		Responder:                   nil,
		RequirePasswordConfirmation: nil,
		handler:                     nil,
		pool:                        nil,
		valkey:                      nil,
	}

	err = app.connect(ctx, config)
	if err != nil {
		return nil, errors.Join(err, app.Close())
	}

	err = app.route(config, logger, assets)
	if err != nil {
		return nil, errors.Join(err, app.Close())
	}

	return app, nil
}

func (a *WebApp) Handler() http.Handler {
	return a.handler
}

func (a *WebApp) Close() error {
	var err error

	if a.valkey != nil {
		err = a.valkey.Close()
	}

	if a.pool != nil {
		a.pool.Close()
	}

	if err != nil {
		return fmt.Errorf("close the web app: %w", err)
	}

	return nil
}

func (a *WebApp) connect(ctx context.Context, config Config) error {
	pool, err := postgres.Open(ctx, webPostgresConfig(config.DB))
	if err != nil {
		return fmt.Errorf("connect to the database: %w", err)
	}

	a.pool = pool
	a.Jobs = newJobQueue(pool)
	a.Users = newUserStore(pool)
	a.RememberTokens = newRememberTokenStore(pool)

	client, err := valkey.Open(valkeyConfig(config.Valkey))
	if err != nil {
		return fmt.Errorf("connect to valkey: %w", err)
	}

	a.valkey = client

	return nil
}

func (a *WebApp) route(config Config, logger *slog.Logger, assets *asset.Server) error {
	session := middleware.Session{}
	themes := middleware.NewThemeCookie(config.App.servedOverHTTPS())
	a.Responder = web.NewResponder(
		web.NewShellBuilder(config.App.Name, assets, session, middleware.AuthenticatedUser{}, themes),
		session,
		logreport.New(logger),
		config.App.Env == EnvironmentLocal,
	)

	auth, err := a.newAuthRoutes(config)
	if err != nil {
		return err
	}

	a.RequirePasswordConfirmation = auth.requirePasswordConfirmation
	a.Mux = newWebMux(webRoutes{
		assets:    assets,
		responder: a.Responder,
		readiness: healthhandler.NewReadiness(
			healthhandler.NewCachedPing(
				healthhandler.NewDependencies(a.pool, a.valkey),
				config.HTTP.ReadinessCacheTTL,
			),
		),
		auth:       auth,
		appearance: settingshandler.NewAppearance(a.Responder, themes),
	})

	chain, err := webMiddleware(webMiddlewareDeps{
		config:    config,
		logger:    logger,
		responder: a.Responder,
		sessions:  valkey.NewSessionStore(a.valkey, config.Session.IdleLifetime),
	})
	if err != nil {
		return err
	}

	a.handler = withMiddleware(a.Mux, chain)

	return nil
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
