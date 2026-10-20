package bootstrap

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/DannyFestor/go-template-html/internal/adapter/broadcast"
	"github.com/DannyFestor/go-template-html/internal/adapter/clock"
	"github.com/DannyFestor/go-template-html/internal/adapter/logreport"
	"github.com/DannyFestor/go-template-html/internal/adapter/valkey"
	"github.com/DannyFestor/go-template-html/internal/adapter/websocket"
	"github.com/DannyFestor/go-template-html/internal/domain"
	realtimehandler "github.com/DannyFestor/go-template-html/internal/handler/realtime"
	"github.com/DannyFestor/go-template-html/internal/middleware"
	"github.com/DannyFestor/go-template-html/internal/view/realtime"
)

const (
	realtimePingInterval = 30 * time.Second
	realtimeWriteTimeout = 10 * time.Second
	realtimeMaxChannels  = 20
	realtimeBufferSize   = 16
)

type realtimeModule struct {
	hub           *realtimehandler.Hub
	socket        *realtimehandler.Socket[*websocket.Conn]
	announcements *broadcast.Broadcaster[domain.Announcement]
	userNotices   *broadcast.Broadcaster[domain.UserNotice]
	stopListening context.CancelFunc
	listening     chan struct{}
	drainTimeout  time.Duration
}

func (a *WebApp) startRealtime(
	ctx context.Context,
	config Config,
	logger *slog.Logger,
	authenticator *middleware.Authenticator,
) {
	a.realtime = newRealtimeModule(ctx, config, logger, a.valkey, authenticator)
	a.Announcements = a.realtime.announcements
	a.UserNotices = a.realtime.userNotices
}

func newRealtimeModule(
	ctx context.Context,
	config Config,
	logger *slog.Logger,
	client *valkey.Client,
	authenticator *middleware.Authenticator,
) *realtimeModule {
	hub := realtimehandler.NewHub(realtimeBufferSize)
	channels := valkey.NewRealtimeChannels(client)
	listenCtx, stopListening := context.WithCancel(ctx)
	module := &realtimeModule{
		hub:           hub,
		socket:        newRealtimeSocket(config, hub, client, authenticator),
		announcements: broadcast.NewBroadcaster(composeAnnouncement, channels, logreport.New(logger)),
		userNotices:   broadcast.NewBroadcaster(composeUserNotice, channels, logreport.New(logger)),
		stopListening: stopListening,
		listening:     make(chan struct{}),
		drainTimeout:  config.App.ShutdownTimeout,
	}

	go module.listen(listenCtx, channels)

	return module
}

func newRealtimeSocket(
	config Config,
	hub *realtimehandler.Hub,
	client *valkey.Client,
	authenticator *middleware.Authenticator,
) *realtimehandler.Socket[*websocket.Conn] {
	return realtimehandler.NewSocket[*websocket.Conn](
		realtimeChannels(),
		hub,
		middleware.NewSessionReader(
			valkey.NewSessionStore(client, config.Session.IdleLifetime),
			clock.System{},
			sessionSettings(config),
			authenticator,
		),
		websocket.NewUpgrader(config.App.URL.Scheme+"://"+config.App.URL.Host),
		realtimehandler.Settings{
			PingInterval: realtimePingInterval,
			WriteTimeout: realtimeWriteTimeout,
			MaxChannels:  realtimeMaxChannels,
		},
	)
}

func realtimeChannels() *realtimehandler.Channels {
	channels := realtimehandler.NewChannels()
	channels.Public(realtime.Announcements)
	channels.Private(realtime.UserChannelPattern, isTheUsersOwnChannel)

	return channels
}

func isTheUsersOwnChannel(user domain.User, params map[string]string) bool {
	return user.ID == domain.UserID(params[realtime.UserChannelParam])
}

func composeAnnouncement(announcement domain.Announcement) broadcast.Message {
	return broadcast.Message{Channel: realtime.Announcements, Fragment: realtime.Banner(announcement.Text)}
}

func composeUserNotice(notice domain.UserNotice) broadcast.Message {
	return broadcast.Message{
		Channel:  realtime.UserChannel(string(notice.UserID)),
		Fragment: realtime.Toast(notice.Text),
	}
}

func (m *realtimeModule) listen(ctx context.Context, channels *valkey.RealtimeChannels) {
	defer close(m.listening)

	channels.Listen(ctx, m.hub.Deliver)
}

func (m *realtimeModule) closeSocketsOnShutdown(server *http.Server) {
	server.RegisterOnShutdown(m.hub.Close)
}

func (m *realtimeModule) close() {
	m.hub.Close()

	select {
	case <-m.hub.Drained():
	case <-time.After(m.drainTimeout):
	}

	m.stopListening()
	<-m.listening
}
