package realtime

import (
	"context"
	"net/http"
)

type conn interface {
	Send(ctx context.Context, payload []byte) error
	Ping(ctx context.Context) error
	Close(code int, reason string) error
	Closed() <-chan struct{}
}

type upgrader[C conn] interface {
	Accept(w http.ResponseWriter, r *http.Request) (C, error)
}
