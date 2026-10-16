package valkey

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	redis     *redis.Client
	keyPrefix string
}

func Open(config Config) (*Client, error) {
	options, err := config.options()
	if err != nil {
		return nil, err
	}

	return &Client{redis: redis.NewClient(options), keyPrefix: config.KeyPrefix}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	err := c.redis.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("ping valkey: %w", err)
	}

	return nil
}

func (c *Client) Close() error {
	err := c.redis.Close()
	if err != nil {
		return fmt.Errorf("close the valkey client: %w", err)
	}

	return nil
}

func (c *Client) key(area, id string) string {
	return c.keyPrefix + area + id
}
