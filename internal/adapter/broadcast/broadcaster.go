package broadcast

import (
	"bytes"
	"context"
	"fmt"
	"io"
)

type component interface {
	Render(ctx context.Context, w io.Writer) error
}

type publisher interface {
	Publish(ctx context.Context, channel string, payload []byte) error
}

type errorReporter interface {
	Report(ctx context.Context, err error)
}

type Message struct {
	Channel  string
	Fragment component
}

type Broadcaster[B any] struct {
	compose   func(B) Message
	publisher publisher
	reporter  errorReporter
}

func NewBroadcaster[B any](compose func(B) Message, publisher publisher, reporter errorReporter) *Broadcaster[B] {
	return &Broadcaster[B]{compose: compose, publisher: publisher, reporter: reporter}
}

func (b *Broadcaster[B]) Broadcast(ctx context.Context, broadcast B) {
	err := b.publish(ctx, b.compose(broadcast))
	if err != nil {
		b.reporter.Report(ctx, err)
	}
}

func (b *Broadcaster[B]) publish(ctx context.Context, message Message) error {
	var rendered bytes.Buffer

	err := message.Fragment.Render(ctx, &rendered)
	if err != nil {
		return fmt.Errorf("render the broadcast on %s: %w", message.Channel, err)
	}

	err = b.publisher.Publish(ctx, message.Channel, rendered.Bytes())
	if err != nil {
		return fmt.Errorf("broadcast on %s: %w", message.Channel, err)
	}

	return nil
}
