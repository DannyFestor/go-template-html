//go:build integration

package postgres_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/aesgcm"
	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/adapter/totp"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

const anyStep domain.TimeStep = 100

type twoFactorFixture struct {
	pool          *postgres.Pool
	store         *postgres.TwoFactorSecretStore
	user          domain.UserID
	secret        domain.TwoFactorSecret
	clock         *testClock
	authenticator *totp.Authenticator
}

func TestAValidCodePassesOnceAndItsReplayFails(t *testing.T) {
	t.Parallel()

	fixture := newTwoFactorFixture(t)
	code := fixture.currentCode()

	assert.True(t, fixture.passes(t, code), "first use")
	assert.False(t, fixture.passes(t, code), "replay")
}

func TestACodeOutsideTheTimeWindowFails(t *testing.T) {
	t.Parallel()

	fixture := newTwoFactorFixture(t)
	stale := fixture.currentCode()

	fixture.clock.advance(2 * time.Minute)

	assert.False(t, fixture.passes(t, stale))
}

func TestACodeOfAnEarlierStepFailsOnceALaterStepWasUsed(t *testing.T) {
	t.Parallel()

	fixture := newTwoFactorFixture(t)
	earlier := fixture.currentCode()

	fixture.clock.advance(30 * time.Second)

	assert.True(t, fixture.passes(t, fixture.currentCode()), "later step")
	assert.False(t, fixture.passes(t, earlier), "earlier step")
}

func TestOnlyOneOfTwoConcurrentUsesOfAStepPasses(t *testing.T) {
	t.Parallel()

	fixture := newTwoFactorFixture(t)
	results := make(chan bool, 2)

	var group sync.WaitGroup
	for range 2 {
		group.Go(func() { results <- useStep(t, fixture.store, fixture.user, anyStep) })
	}

	group.Wait()
	close(results)

	assert.ElementsMatch(t, []bool{true, false}, []bool{<-results, <-results})
}

func TestUsingAStepOfOneUserLeavesItFreeForAnother(t *testing.T) {
	t.Parallel()

	fixture := newTwoFactorFixture(t)
	other := createUser(t, fixture.pool)
	require.NoError(t, fixture.store.Save(t.Context(), other, totp.NewSecret()))

	require.True(t, useStep(t, fixture.store, fixture.user, anyStep))

	assert.True(t, useStep(t, fixture.store, other, anyStep))
}

func TestFindReturnsTheSavedSecret(t *testing.T) {
	t.Parallel()

	fixture := newTwoFactorFixture(t)

	found, ok, err := fixture.store.Find(t.Context(), fixture.user)

	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, fixture.secret, found)
}

func TestFindReportsAUserWithoutASecret(t *testing.T) {
	t.Parallel()

	fixture := newTwoFactorFixture(t)

	_, ok, err := fixture.store.Find(t.Context(), newUserID())

	require.NoError(t, err)
	assert.False(t, ok)
}

func TestTheSecretIsStoredEncryptedUnderTheGivenKey(t *testing.T) {
	t.Parallel()

	fixture := newTwoFactorFixture(t)
	otherKeyStore := postgres.NewTwoFactorSecretStore(fixture.pool, newEncrypter(t, encryptionKeyFill+1))

	_, _, err := otherKeyStore.Find(t.Context(), fixture.user)

	require.ErrorIs(t, err, aesgcm.ErrUndecryptable)
}

func TestSavingAgainReplacesTheSecretAndForgetsItsUsedSteps(t *testing.T) {
	t.Parallel()

	fixture := newTwoFactorFixture(t)
	replacement := totp.NewSecret()

	require.True(t, useStep(t, fixture.store, fixture.user, anyStep))

	require.NoError(t, fixture.store.Save(t.Context(), fixture.user, replacement))

	found, _, err := fixture.store.Find(t.Context(), fixture.user)
	require.NoError(t, err)
	assert.Equal(t, replacement, found)
	assert.True(t, useStep(t, fixture.store, fixture.user, anyStep))
}

func TestDeleteRemovesTheSecret(t *testing.T) {
	t.Parallel()

	fixture := newTwoFactorFixture(t)

	require.NoError(t, fixture.store.Delete(t.Context(), fixture.user))

	_, ok, err := fixture.store.Find(t.Context(), fixture.user)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.False(t, useStep(t, fixture.store, fixture.user, anyStep))
}

func newTwoFactorFixture(t *testing.T) *twoFactorFixture {
	t.Helper()

	pool := openTestDatabase(t)
	clock := clockAt(time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC))
	fixture := &twoFactorFixture{
		pool:          pool,
		store:         postgres.NewTwoFactorSecretStore(pool, newEncrypter(t, encryptionKeyFill)),
		user:          createUser(t, pool),
		secret:        totp.NewSecret(),
		clock:         clock,
		authenticator: totp.NewAuthenticator(clock),
	}

	require.NoError(t, fixture.store.Save(t.Context(), fixture.user, fixture.secret))

	return fixture
}

func (f *twoFactorFixture) currentCode() string {
	return totp.Code(f.secret, f.clock.Now())
}

func (f *twoFactorFixture) passes(t *testing.T, code string) bool {
	t.Helper()

	secret, ok, err := f.store.Find(t.Context(), f.user)
	require.NoError(t, err)
	require.True(t, ok)

	step, matched := f.authenticator.Match(secret, code)

	return matched && useStep(t, f.store, f.user, step)
}

func useStep(t *testing.T, store *postgres.TwoFactorSecretStore, user domain.UserID, step domain.TimeStep) bool {
	t.Helper()

	used, err := store.UseStep(t.Context(), user, step)
	require.NoError(t, err)

	return used
}
