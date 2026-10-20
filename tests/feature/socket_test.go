//go:build feature

package feature_test

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/coder/websocket"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/route"
)

const (
	rebroadcastInterval = 50 * time.Millisecond
	receiveTimeout      = 5 * time.Second
)

type socket struct {
	conn     *websocket.Conn
	received chan []byte
	closed   chan error
}

func (a *app) dial(t *testing.T, channels ...string) (*socket, int) {
	t.Helper()

	return a.dialFrom(t, nil, channels...)
}

func (a *app) dialFrom(t *testing.T, header http.Header, channels ...string) (*socket, int) {
	t.Helper()

	url := "ws" + strings.TrimPrefix(a.server.URL, "http") + route.Realtime + "?channels=" + strings.Join(channels, ",")

	conn, handshake, err := websocket.Dial(t.Context(), url, &websocket.DialOptions{
		HTTPClient:           a.server.Client(),
		HTTPHeader:           header,
		Host:                 "",
		Subprotocols:         nil,
		CompressionMode:      websocket.CompressionDisabled,
		CompressionThreshold: 0,
		OnPingReceived:       nil,
		OnPongReceived:       nil,
	})
	require.NotNil(t, handshake, "the upgrade failed without a response: %v", err)

	if handshake.Body != nil {
		require.NoError(t, handshake.Body.Close())
	}

	if err != nil {
		return nil, handshake.StatusCode
	}

	return openSocket(t, conn), handshake.StatusCode
}

func openSocket(t *testing.T, conn *websocket.Conn) *socket {
	t.Helper()

	t.Cleanup(func() { _ = conn.CloseNow() })

	opened := &socket{conn: conn, received: make(chan []byte), closed: make(chan error, 1)}

	go opened.receive(t.Context())

	return opened
}

func (a *app) mustDial(t *testing.T, channels ...string) *socket {
	t.Helper()

	opened, status := a.dial(t, channels...)
	require.Equal(t, http.StatusSwitchingProtocols, status)

	return opened
}

func (s *socket) receive(ctx context.Context) {
	for {
		_, payload, err := s.conn.Read(ctx)
		if err != nil {
			s.closed <- err

			return
		}

		select {
		case s.received <- payload:
		case <-ctx.Done():
			return
		}
	}
}

func (s *socket) receiveWhileBroadcasting(t *testing.T, broadcast func()) *goquery.Document {
	t.Helper()

	ticker := time.NewTicker(rebroadcastInterval)
	defer ticker.Stop()

	deadline := time.After(receiveTimeout)

	for {
		broadcast()

		select {
		case payload := <-s.received:
			return parseFragment(t, payload)
		case err := <-s.closed:
			require.FailNow(t, "the socket closed before the broadcast arrived", err)
		case <-deadline:
			require.FailNow(t, "the broadcast never arrived")
		case <-ticker.C:
		}
	}
}

func (s *socket) closeStatus(t *testing.T) websocket.StatusCode {
	t.Helper()

	select {
	case err := <-s.closed:
		return websocket.CloseStatus(err)
	case <-time.After(receiveTimeout):
		require.FailNow(t, "the socket stayed open")

		return 0
	}
}

func parseFragment(t *testing.T, payload []byte) *goquery.Document {
	t.Helper()

	document, err := goquery.NewDocumentFromReader(bytes.NewReader(payload))
	require.NoError(t, err)

	return document
}
