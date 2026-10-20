package websocket_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	coderws "github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/websocket"
)

const (
	appOrigin     = "https://app.example"
	shortTimeout  = 100 * time.Millisecond
	waitForServer = 5 * time.Second
)

type accepted struct {
	conn *websocket.Conn
	err  error
}

func serveUpgrades(t *testing.T, serverTimeout time.Duration) (string, <-chan accepted) {
	t.Helper()

	conns := make(chan accepted, 1)
	upgrader := websocket.NewUpgrader(appOrigin)

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Accept(w, r)
		conns <- accepted{conn: conn, err: err}

		if err == nil {
			<-conn.Closed()
		}
	}))
	server.Config.ReadTimeout = serverTimeout
	server.Config.WriteTimeout = serverTimeout
	server.Start()
	t.Cleanup(server.Close)

	return "ws" + strings.TrimPrefix(server.URL, "http"), conns
}

func dial(t *testing.T, url string, header http.Header) (*coderws.Conn, int, error) {
	t.Helper()

	conn, handshake, err := coderws.Dial(t.Context(), url, &coderws.DialOptions{
		HTTPClient:           nil,
		HTTPHeader:           header,
		Host:                 "",
		Subprotocols:         nil,
		CompressionMode:      coderws.CompressionDisabled,
		CompressionThreshold: 0,
		OnPingReceived:       nil,
		OnPongReceived:       nil,
	})
	if err == nil {
		t.Cleanup(func() { _ = conn.CloseNow() })
	}

	require.NotNil(t, handshake, "the upgrade failed without a response: %v", err)

	if handshake.Body != nil {
		require.NoError(t, handshake.Body.Close())
	}

	return conn, handshake.StatusCode, err //nolint:wrapcheck // Tests assert on the result itself.
}

func acceptOne(t *testing.T, serverTimeout time.Duration) (*coderws.Conn, *websocket.Conn) {
	t.Helper()

	url, conns := serveUpgrades(t, serverTimeout)
	client, _, err := dial(t, url, nil)
	require.NoError(t, err)

	server := <-conns
	require.NoError(t, server.err)

	return client, server.conn
}

func TestASentMessageReachesTheClientAsText(t *testing.T) {
	t.Parallel()

	client, server := acceptOne(t, time.Minute)

	require.NoError(t, server.Send(t.Context(), []byte("<div>hello</div>")))

	kind, payload, err := client.Read(t.Context())
	require.NoError(t, err)
	assert.Equal(t, coderws.MessageText, kind)
	assert.Equal(t, "<div>hello</div>", string(payload))
}

func TestAnInboundMessageClosesTheConnectionAsAPolicyViolation(t *testing.T) {
	t.Parallel()

	client, _ := acceptOne(t, time.Minute)

	require.NoError(t, client.Write(t.Context(), coderws.MessageText, []byte("hi")))

	_, _, err := client.Read(t.Context())
	assert.Equal(t, coderws.StatusPolicyViolation, coderws.CloseStatus(err))
}

func TestClosedReportsTheClientClosingTheConnection(t *testing.T) {
	t.Parallel()

	client, server := acceptOne(t, time.Minute)

	require.NoError(t, client.Close(coderws.StatusNormalClosure, "bye"))

	requireClosed(t, server)
}

func TestCloseSendsTheCloseCodeToTheClient(t *testing.T) {
	t.Parallel()

	client, server := acceptOne(t, time.Minute)

	go func() { _ = server.Close(int(coderws.StatusGoingAway), "shutting down") }()

	_, _, err := client.Read(t.Context())
	assert.Equal(t, coderws.StatusGoingAway, coderws.CloseStatus(err))
}

func TestPingIsAnsweredByTheClient(t *testing.T) {
	t.Parallel()

	client, server := acceptOne(t, time.Minute)
	go func() { _, _, _ = client.Read(t.Context()) }()

	assert.NoError(t, server.Ping(t.Context()))
}

func TestTheConnectionOutlivesTheServersReadAndWriteTimeouts(t *testing.T) {
	t.Parallel()

	client, server := acceptOne(t, shortTimeout)

	time.Sleep(3 * shortTimeout)

	require.NoError(t, server.Send(t.Context(), []byte("still here")))
	_, payload, err := client.Read(t.Context())
	require.NoError(t, err)
	assert.Equal(t, "still here", string(payload))
}

func TestTheOriginMustBeTheRequestHostOrTheAppURL(t *testing.T) {
	t.Parallel()

	for origin, wantStatus := range map[string]int{
		appOrigin:              http.StatusSwitchingProtocols,
		"https://evil.example": http.StatusForbidden,
		"http://app.example":   http.StatusForbidden,
	} {
		t.Run(origin, func(t *testing.T) {
			t.Parallel()

			url, _ := serveUpgrades(t, time.Minute)

			_, status, _ := dial(t, url, http.Header{"Origin": {origin}})

			assert.Equal(t, wantStatus, status)
		})
	}
}

func requireClosed(t *testing.T, server *websocket.Conn) {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), waitForServer)
	defer cancel()

	select {
	case <-server.Closed():
	case <-ctx.Done():
		require.FailNow(t, "the server side stayed open")
	}
}
