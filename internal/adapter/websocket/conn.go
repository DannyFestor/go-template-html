package websocket

import (
	"context"
	"fmt"

	"github.com/coder/websocket"
)

type Conn struct {
	conn       *websocket.Conn
	readClosed context.Context //nolint:containedctx // The library reports the closed read side only as a context.
}

func (c *Conn) Send(ctx context.Context, payload []byte) error {
	err := c.conn.Write(ctx, websocket.MessageText, payload)
	if err != nil {
		return fmt.Errorf("send on the websocket: %w", err)
	}

	return nil
}

func (c *Conn) Ping(ctx context.Context) error {
	err := c.conn.Ping(ctx)
	if err != nil {
		return fmt.Errorf("ping the websocket: %w", err)
	}

	return nil
}

func (c *Conn) Close(code int, reason string) error {
	err := c.conn.Close(websocket.StatusCode(code), reason)
	if err != nil {
		return fmt.Errorf("close the websocket with %d: %w", code, err)
	}

	return nil
}

func (c *Conn) Closed() <-chan struct{} {
	return c.readClosed.Done()
}
