package realtime

import (
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/DannyFestor/go-template-html/internal/web"
)

const (
	channelsParam     = "channels"
	channelsSeparator = ","
)

type Settings struct {
	PingInterval time.Duration
	WriteTimeout time.Duration
	MaxChannels  int
}

type Socket[C conn] struct {
	channels *Channels
	hub      *Hub
	users    sessionUsers
	upgrader upgrader[C]
	settings Settings
}

func NewSocket[C conn](
	channels *Channels,
	hub *Hub,
	users sessionUsers,
	upgrader upgrader[C],
	settings Settings,
) *Socket[C] {
	return &Socket[C]{channels: channels, hub: hub, users: users, upgrader: upgrader, settings: settings}
}

func (s *Socket[C]) Serve(w http.ResponseWriter, r *http.Request) error {
	channels, err := s.requestedChannels(r)
	if err != nil {
		return err
	}

	err = s.authorize(r, channels)
	if err != nil {
		return err
	}

	subscribed := s.hub.subscribe(channels)
	defer s.hub.unsubscribe(subscribed)

	s.streamUnlessTheUpgraderRefused(w, r, subscribed)

	return nil
}

func (s *Socket[C]) streamUnlessTheUpgraderRefused(w http.ResponseWriter, r *http.Request, subscribed *subscription) {
	upgraded, err := s.upgrader.Accept(w, r)
	if err == nil {
		s.stream(r, upgraded, subscribed)
	}
}

func (s *Socket[C]) requestedChannels(r *http.Request) ([]string, error) {
	channels := strings.Split(r.URL.Query().Get(channelsParam), channelsSeparator)
	if len(channels) > s.settings.MaxChannels || slices.Contains(channels, "") {
		return nil, web.ErrBadRequest
	}

	slices.Sort(channels)

	return slices.Compact(channels), nil
}

func (s *Socket[C]) authorize(r *http.Request, channels []string) error {
	allowed, err := s.channels.allowAll(channels, newLazyUser(s.users, r))
	if err != nil {
		return err
	}

	if !allowed {
		return web.ErrForbidden
	}

	return nil
}
