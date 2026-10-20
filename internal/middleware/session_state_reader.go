package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type sessionFetcher func(ctx context.Context, id string) ([]byte, bool, error)

type sessionStateReader struct {
	cookie hostOnlyCookie
	clock  clock
	fetch  sessionFetcher
}

func (s sessionStateReader) read(r *http.Request) (*sessionState, error) {
	value, found := s.cookie.read(r)
	if !found {
		return guestSessionState(), nil
	}

	token := sessionToken(value)

	encoded, stored, err := s.fetch(r.Context(), token.storageID())
	if err != nil {
		return nil, fmt.Errorf("load the session: %w", err)
	}

	if !stored {
		return staleSessionState(nil), nil
	}

	var data sessionData

	err = json.Unmarshal(encoded, &data)
	if err != nil {
		return nil, fmt.Errorf("decode the session: %w", err)
	}

	if !s.clock.Now().Before(data.ExpiresAt) {
		return staleSessionState([]sessionToken{token}), nil
	}

	return loadedSessionState(token, data), nil
}
