package websocket

import (
	"fmt"
	"net/http"

	"github.com/coder/websocket"
)

const readLimitBytes = 4 << 10

type Upgrader struct {
	appOrigin string
}

func NewUpgrader(appOrigin string) *Upgrader {
	return &Upgrader{appOrigin: appOrigin}
}

func (u *Upgrader) Accept(w http.ResponseWriter, r *http.Request) (*Conn, error) {
	conn, err := websocket.Accept(w, r, u.options())
	if err != nil {
		return nil, fmt.Errorf("accept the websocket: %w", err)
	}

	conn.SetReadLimit(readLimitBytes)

	return &Conn{conn: conn, readClosed: conn.CloseRead(r.Context())}, nil
}

func (u *Upgrader) options() *websocket.AcceptOptions {
	return &websocket.AcceptOptions{
		Subprotocols:         nil,
		InsecureSkipVerify:   false,
		OriginPatterns:       []string{u.appOrigin},
		CompressionMode:      websocket.CompressionDisabled,
		CompressionThreshold: 0,
		OnPingReceived:       nil,
		OnPongReceived:       nil,
	}
}
