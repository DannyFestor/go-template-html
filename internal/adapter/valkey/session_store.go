package valkey

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const sessionArea = "session:"

type SessionStore struct {
	client *Client
	idle   time.Duration
}

func NewSessionStore(client *Client, idle time.Duration) *SessionStore {
	return &SessionStore{client: client, idle: idle}
}

// Load restarts the idle lifetime in the same command, so a session in use never needs a write to stay alive.
func (s *SessionStore) Load(ctx context.Context, id string) ([]byte, bool, error) {
	return storedSession(s.client.redis.GetEx(ctx, s.client.key(sessionArea, id), s.idle).Bytes())
}

func (s *SessionStore) Peek(ctx context.Context, id string) ([]byte, bool, error) {
	return storedSession(s.client.redis.Get(ctx, s.client.key(sessionArea, id)).Bytes())
}

func storedSession(data []byte, err error) ([]byte, bool, error) {
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, fmt.Errorf("load the session: %w", err)
	}

	return data, true, nil
}

func (s *SessionStore) Save(ctx context.Context, id string, data []byte) error {
	err := s.client.redis.Set(ctx, s.client.key(sessionArea, id), data, s.idle).Err()
	if err != nil {
		return fmt.Errorf("save the session: %w", err)
	}

	return nil
}

func (s *SessionStore) Delete(ctx context.Context, id string) error {
	err := s.client.redis.Del(ctx, s.client.key(sessionArea, id)).Err()
	if err != nil {
		return fmt.Errorf("delete the session: %w", err)
	}

	return nil
}
