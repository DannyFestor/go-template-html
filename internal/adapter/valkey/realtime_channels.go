package valkey

import (
	"context"
	"fmt"
	"strings"
)

const (
	realtimeArea   = "realtime:"
	anyRealtimeKey = "*"
)

type RealtimeChannels struct {
	client *Client
}

func NewRealtimeChannels(client *Client) *RealtimeChannels {
	return &RealtimeChannels{client: client}
}

func (c *RealtimeChannels) Publish(ctx context.Context, channel string, payload []byte) error {
	err := c.client.redis.Publish(ctx, c.client.key(realtimeArea, channel), payload).Err()
	if err != nil {
		return fmt.Errorf("publish on %s: %w", channel, err)
	}

	return nil
}

func (c *RealtimeChannels) Listen(ctx context.Context, deliver func(channel string, payload []byte)) {
	subscription := c.client.redis.PSubscribe(ctx, c.client.key(realtimeArea, anyRealtimeKey))
	defer func() { _ = subscription.Close() }()

	messages := subscription.Channel()
	prefix := c.client.key(realtimeArea, "")

	for {
		select {
		case <-ctx.Done():
			return
		case message := <-messages:
			deliver(strings.TrimPrefix(message.Channel, prefix), []byte(message.Payload))
		}
	}
}
