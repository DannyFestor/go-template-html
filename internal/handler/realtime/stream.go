package realtime

import (
	"context"
	"errors"
	"net/http"
	"time"
)

const (
	closeGoingAway       = 1001
	closePolicyViolation = 1008
	closeInternalError   = 1011
)

func (s *Socket[C]) stream(r *http.Request, upgraded C, subscribed *subscription) {
	ticker := time.NewTicker(s.settings.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case payload := <-subscribed.messages:
			if !s.send(r.Context(), upgraded, payload) {
				return
			}
		case <-subscribed.ended:
			_ = upgraded.Close(closeCodeFor(subscribed.reason), subscribed.reason.Error())

			return
		case <-ticker.C:
			if !s.keepAlive(r, upgraded, subscribed.channels) {
				return
			}
		case <-upgraded.Closed():
			return
		}
	}
}

func (s *Socket[C]) send(ctx context.Context, upgraded C, payload []byte) bool {
	ctx, cancel := context.WithTimeout(ctx, s.settings.WriteTimeout)
	defer cancel()

	return upgraded.Send(ctx, payload) == nil
}

func (s *Socket[C]) keepAlive(r *http.Request, upgraded C, channels []string) bool {
	allowed, err := s.channels.allowAll(channels, newLazyUser(s.users, r))
	if err != nil {
		_ = upgraded.Close(closeInternalError, "the session could not be checked")

		return false
	}

	if !allowed {
		_ = upgraded.Close(closePolicyViolation, "the session ended")

		return false
	}

	return s.ping(r.Context(), upgraded)
}

func (s *Socket[C]) ping(ctx context.Context, upgraded C) bool {
	ctx, cancel := context.WithTimeout(ctx, s.settings.WriteTimeout)
	defer cancel()

	return upgraded.Ping(ctx) == nil
}

func closeCodeFor(reason error) int {
	if errors.Is(reason, errHubClosed) {
		return closeGoingAway
	}

	return closePolicyViolation
}
