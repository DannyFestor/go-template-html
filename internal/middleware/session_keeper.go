package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type sessionStore interface {
	Load(ctx context.Context, id string) ([]byte, bool, error)
	Save(ctx context.Context, id string, data []byte) error
	Delete(ctx context.Context, id string) error
}

type clock interface {
	Now() time.Time
}

type sessionKeeper struct {
	store            sessionStore
	clock            clock
	cookie           hostOnlyCookie
	absoluteLifetime time.Duration
}

func (k *sessionKeeper) load(r *http.Request) (*sessionState, error) {
	value, found := k.cookie.read(r)
	if !found {
		return guestSessionState(), nil
	}

	token := sessionToken(value)

	encoded, stored, err := k.store.Load(r.Context(), token.storageID())
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

	if !k.clock.Now().Before(data.ExpiresAt) {
		return staleSessionState([]sessionToken{token}), nil
	}

	return loadedSessionState(token, data), nil
}

func (k *sessionKeeper) commit(ctx context.Context, w http.ResponseWriter, state *sessionState) error {
	if !state.changed {
		return nil
	}

	if state.data.isEmpty() {
		return k.discard(ctx, w, state)
	}

	minted := state.renew || state.token == ""
	if minted {
		k.mint(state)
	}

	err := k.save(ctx, state)
	if err != nil {
		return err
	}

	if minted {
		k.cookie.set(w, string(state.token), 0)
	}

	return nil
}

func (k *sessionKeeper) discard(ctx context.Context, w http.ResponseWriter, state *sessionState) error {
	state.retireToken()

	err := k.deleteRetired(ctx, state)
	if err != nil {
		return err
	}

	if state.hadCookie {
		k.cookie.clear(w)
	}

	return nil
}

func (k *sessionKeeper) mint(state *sessionState) {
	state.retireToken()
	state.token = newSessionToken()
	state.data.ExpiresAt = k.clock.Now().Add(k.absoluteLifetime)
}

func (k *sessionKeeper) save(ctx context.Context, state *sessionState) error {
	err := k.deleteRetired(ctx, state)
	if err != nil {
		return err
	}

	encoded, err := json.Marshal(state.data)
	if err != nil {
		return fmt.Errorf("encode the session: %w", err)
	}

	err = k.store.Save(ctx, state.token.storageID(), encoded)
	if err != nil {
		return fmt.Errorf("save the session: %w", err)
	}

	return nil
}

func (k *sessionKeeper) deleteRetired(ctx context.Context, state *sessionState) error {
	for _, token := range state.retired {
		err := k.store.Delete(ctx, token.storageID())
		if err != nil {
			return fmt.Errorf("delete the retired session: %w", err)
		}
	}

	return nil
}
