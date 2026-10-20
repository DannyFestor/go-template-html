package discard

import "context"

type Broadcaster[B any] struct{}

func (Broadcaster[B]) Broadcast(context.Context, B) {}
