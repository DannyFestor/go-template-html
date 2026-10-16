//go:build integration

package valkey_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/valkey"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testvalkey"
)

const (
	sessionID = "6b86b273ff34fce19d6b804eff5a3f57"
	payload   = `{"user_id":"42"}`
)

func TestASavedSessionLoadsBack(t *testing.T) {
	t.Parallel()

	store := newSessionStore(t, testvalkey.New(t), time.Hour)
	require.NoError(t, store.Save(t.Context(), sessionID, []byte(payload)))

	loaded, found, err := store.Load(t.Context(), sessionID)

	require.NoError(t, err)
	assert.True(t, found)
	assert.JSONEq(t, payload, string(loaded))
}

func TestAnUnknownSessionIsNotFound(t *testing.T) {
	t.Parallel()

	_, found, err := newSessionStore(t, testvalkey.New(t), time.Hour).Load(t.Context(), sessionID)

	require.NoError(t, err)
	assert.False(t, found)
}

func TestADeletedSessionIsNotFound(t *testing.T) {
	t.Parallel()

	store := newSessionStore(t, testvalkey.New(t), time.Hour)
	require.NoError(t, store.Save(t.Context(), sessionID, []byte(payload)))
	require.NoError(t, store.Delete(t.Context(), sessionID))

	_, found, err := store.Load(t.Context(), sessionID)

	require.NoError(t, err)
	assert.False(t, found)
}

func TestASessionExpiresOnceIdleForItsLifetime(t *testing.T) {
	t.Parallel()

	store := newSessionStore(t, testvalkey.New(t), 300*time.Millisecond)
	require.NoError(t, store.Save(t.Context(), sessionID, []byte(payload)))

	time.Sleep(500 * time.Millisecond)

	_, found, err := store.Load(t.Context(), sessionID)

	require.NoError(t, err)
	assert.False(t, found)
}

func TestLoadingASessionRestartsItsIdleLifetime(t *testing.T) {
	t.Parallel()

	store := newSessionStore(t, testvalkey.New(t), 600*time.Millisecond)
	require.NoError(t, store.Save(t.Context(), sessionID, []byte(payload)))

	for range 4 {
		time.Sleep(300 * time.Millisecond)

		_, found, err := store.Load(t.Context(), sessionID)
		require.NoError(t, err)
		require.True(t, found, "expired although loaded within its idle lifetime")
	}
}

func TestSessionsAreIsolatedByKeyPrefix(t *testing.T) {
	t.Parallel()

	first, second := testvalkey.New(t), testvalkey.New(t)
	require.NoError(t, newSessionStore(t, first, time.Hour).Save(t.Context(), sessionID, []byte(payload)))

	_, found, err := newSessionStore(t, second, time.Hour).Load(t.Context(), sessionID)

	require.NoError(t, err)
	assert.False(t, found)
}

func TestLoadingFailsWhileValkeyIsUnreachable(t *testing.T) {
	t.Parallel()

	_, _, err := newSessionStore(t, testvalkey.Unreachable(t), time.Hour).Load(t.Context(), sessionID)

	assert.Error(t, err)
}

func newSessionStore(t *testing.T, instance testvalkey.Instance, idle time.Duration) *valkey.SessionStore {
	t.Helper()

	return valkey.NewSessionStore(open(t, instance), idle)
}

func open(t *testing.T, instance testvalkey.Instance) *valkey.Client {
	t.Helper()

	client, err := valkey.Open(instance.Config())
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, client.Close()) })

	return client
}
