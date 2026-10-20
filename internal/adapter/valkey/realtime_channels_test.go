//go:build integration

package valkey_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/valkey"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testvalkey"
)

const (
	republishInterval = 20 * time.Millisecond
	deliveryTimeout   = 5 * time.Second
)

type delivery struct {
	channel string
	payload string
}

func listen(t *testing.T, channels *valkey.RealtimeChannels) (<-chan delivery, <-chan struct{}, context.CancelFunc) {
	t.Helper()

	ctx, cancel := context.WithCancel(t.Context())
	delivered := make(chan delivery, 16)
	stopped := make(chan struct{})

	go func() {
		defer close(stopped)

		channels.Listen(ctx, func(channel string, payload []byte) {
			delivered <- delivery{channel: channel, payload: string(payload)}
		})
	}()

	t.Cleanup(cancel)

	return delivered, stopped, cancel
}

func publishUntilDelivered(
	t *testing.T,
	channels *valkey.RealtimeChannels,
	delivered <-chan delivery,
	channel string,
) delivery {
	t.Helper()

	deadline := time.After(deliveryTimeout)

	for {
		require.NoError(t, channels.Publish(t.Context(), channel, []byte("<div>hello</div>")))

		select {
		case received := <-delivered:
			return received
		case <-deadline:
			require.FailNow(t, "nothing was delivered")
		case <-time.After(republishInterval):
		}
	}
}

func TestAPublishedPayloadIsDeliveredUnderItsChannelName(t *testing.T) {
	t.Parallel()

	channels := valkey.NewRealtimeChannels(open(t, testvalkey.New(t)))
	delivered, _, _ := listen(t, channels)

	received := publishUntilDelivered(t, channels, delivered, "users.ada")

	assert.Equal(t, delivery{channel: "users.ada", payload: "<div>hello</div>"}, received)
}

func TestRealtimeChannelsAreIsolatedByKeyPrefix(t *testing.T) {
	t.Parallel()

	listening := valkey.NewRealtimeChannels(open(t, testvalkey.New(t)))
	other := valkey.NewRealtimeChannels(open(t, testvalkey.New(t)))
	delivered, _, _ := listen(t, listening)
	publishUntilDelivered(t, listening, delivered, "announcements")

	require.NoError(t, other.Publish(t.Context(), "announcements", []byte("elsewhere")))
	require.NoError(t, listening.Publish(t.Context(), "announcements", []byte("here")))

	for received := range delivered {
		if received.payload != "<div>hello</div>" {
			assert.Equal(t, "here", received.payload)

			return
		}
	}
}

func TestListeningStopsWhenItsContextEnds(t *testing.T) {
	t.Parallel()

	_, stopped, cancel := listen(t, valkey.NewRealtimeChannels(open(t, testvalkey.New(t))))

	cancel()

	select {
	case <-stopped:
	case <-time.After(deliveryTimeout):
		require.FailNow(t, "the listener kept running")
	}
}

func TestPublishingFailsWhileValkeyIsUnreachable(t *testing.T) {
	t.Parallel()

	channels := valkey.NewRealtimeChannels(open(t, testvalkey.Unreachable(t)))

	assert.Error(t, channels.Publish(t.Context(), "announcements", []byte("lost")))
}
