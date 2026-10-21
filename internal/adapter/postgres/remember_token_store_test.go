//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/adapter/remembertoken"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

const replayWindow = 30 * time.Second

type rememberTokenFixture struct {
	pool   *postgres.Pool
	store  *postgres.RememberTokenStore
	userID domain.UserID
}

func newRememberTokenFixture(t *testing.T) rememberTokenFixture {
	t.Helper()

	pool := openTestDatabase(t)

	return rememberTokenFixture{
		pool:   pool,
		store:  postgres.NewRememberTokenStore(pool, replayWindow),
		userID: createUser(t, pool),
	}
}

func (f rememberTokenFixture) issue(t *testing.T, expiresAt time.Time) domain.RememberToken {
	t.Helper()

	token := remembertoken.Generator{}.New()
	require.NoError(t, f.store.Issue(t.Context(), f.userID, token, expiresAt))

	return token
}

func (f rememberTokenFixture) consume(
	t *testing.T,
	token domain.RememberToken,
	now time.Time,
) (domain.RememberTokenUse, bool) {
	t.Helper()

	use, consumed, err := f.store.Consume(t.Context(), token, now)
	require.NoError(t, err)

	return use, consumed
}

func noon() time.Time {
	return time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
}

func TestAnIssuedRememberTokenIsConsumedForItsUser(t *testing.T) {
	t.Parallel()

	fixture := newRememberTokenFixture(t)
	token := fixture.issue(t, noon().Add(time.Hour))

	use, consumed := fixture.consume(t, token, noon())

	assert.True(t, consumed)
	assert.Equal(t, domain.RememberTokenUse{UserID: fixture.userID, Replay: false}, use)
}

func TestAConsumedRememberTokenIsAReplayUntilTheReplayWindowCloses(t *testing.T) {
	t.Parallel()

	fixture := newRememberTokenFixture(t)
	token := fixture.issue(t, noon().Add(time.Hour))
	fixture.consume(t, token, noon())

	replayed, consumedAgain := fixture.consume(t, token, noon().Add(replayWindow-time.Second))
	assert.True(t, consumedAgain)
	assert.Equal(t, domain.RememberTokenUse{UserID: fixture.userID, Replay: true}, replayed)

	_, consumedAfterTheWindow := fixture.consume(t, token, noon().Add(replayWindow))
	assert.False(t, consumedAfterTheWindow)
}

func TestAConcurrentConsumeWaitsForTheFirstAndIsAReplay(t *testing.T) {
	t.Parallel()

	fixture := newRememberTokenFixture(t)
	token := fixture.issue(t, noon().Add(time.Hour))
	firstConsumed, releaseFirst := make(chan struct{}), make(chan struct{})
	firstDone := make(chan error, 1)

	go func() {
		firstDone <- postgres.NewTransactor(fixture.pool).WithinTransaction(t.Context(), func(ctx context.Context) error {
			_, _, err := fixture.store.Consume(ctx, token, noon())

			close(firstConsumed)
			<-releaseFirst

			return err //nolint:wrapcheck // The test only checks that the first consume succeeded.
		})
	}()

	<-firstConsumed
	time.AfterFunc(100*time.Millisecond, func() { close(releaseFirst) })

	second, consumed := fixture.consume(t, token, noon())

	require.NoError(t, <-firstDone)
	assert.True(t, consumed)
	assert.True(t, second.Replay)
}

func TestAReplayDoesNotExtendTheReplayWindow(t *testing.T) {
	t.Parallel()

	fixture := newRememberTokenFixture(t)
	token := fixture.issue(t, noon().Add(time.Hour))
	fixture.consume(t, token, noon())
	fixture.consume(t, token, noon().Add(replayWindow/2))

	_, consumed := fixture.consume(t, token, noon().Add(replayWindow+time.Second))

	assert.False(t, consumed)
}

func TestARememberTokenWithAWrongValidatorIsNotConsumed(t *testing.T) {
	t.Parallel()

	fixture := newRememberTokenFixture(t)
	token := fixture.issue(t, noon().Add(time.Hour))
	forged := domain.RememberToken{Selector: token.Selector, Validator: token.Validator + "X"}

	_, consumed := fixture.consume(t, forged, noon())
	assert.False(t, consumed)

	use, consumedWithTheRightValidator := fixture.consume(t, token, noon())
	assert.True(t, consumedWithTheRightValidator)
	assert.False(t, use.Replay)
}

func TestARememberTokenIsNotConsumedOnceItExpires(t *testing.T) {
	t.Parallel()

	fixture := newRememberTokenFixture(t)
	token := fixture.issue(t, noon())

	_, consumed := fixture.consume(t, token, noon())

	assert.False(t, consumed)
}

func TestAForgottenRememberTokenIsNotConsumed(t *testing.T) {
	t.Parallel()

	fixture := newRememberTokenFixture(t)
	token := fixture.issue(t, noon().Add(time.Hour))

	require.NoError(t, fixture.store.Forget(t.Context(), token))

	_, consumed := fixture.consume(t, token, noon())
	assert.False(t, consumed)
}

func TestForgettingTheDevicesOfAUserForgetsAllItsRememberTokensButNoOneElses(t *testing.T) {
	t.Parallel()

	fixture := newRememberTokenFixture(t)
	first := fixture.issue(t, noon().Add(time.Hour))
	second := fixture.issue(t, noon().Add(time.Hour))
	othersToken := remembertoken.Generator{}.New()
	require.NoError(
		t,
		fixture.store.Issue(t.Context(), createUser(t, fixture.pool), othersToken, noon().Add(time.Hour)),
	)

	require.NoError(t, fixture.store.ForgetDevicesOf(t.Context(), fixture.userID))

	for token, kept := range map[domain.RememberToken]bool{first: false, second: false, othersToken: true} {
		_, found := fixture.consume(t, token, noon())
		assert.Equal(t, kept, found, token.Selector)
	}
}

func TestDeleteExpiredDeletesTheExpiredTokensAndTheReplacedOnesPastTheReplayWindow(t *testing.T) {
	t.Parallel()

	fixture := newRememberTokenFixture(t)
	expired := fixture.issue(t, noon())
	live := fixture.issue(t, noon().Add(time.Hour))
	replacedLongAgo := fixture.issue(t, noon().Add(time.Hour))
	replacedJustNow := fixture.issue(t, noon().Add(time.Hour))
	fixture.consume(t, replacedLongAgo, noon().Add(-replayWindow))
	fixture.consume(t, replacedJustNow, noon().Add(-replayWindow+time.Second))

	require.NoError(t, fixture.store.DeleteExpired(t.Context(), noon()))

	var beforeAnyExpiry time.Time

	for token, kept := range map[domain.RememberToken]bool{
		expired:         false,
		live:            true,
		replacedLongAgo: false,
		replacedJustNow: true,
	} {
		_, found := fixture.consume(t, token, beforeAnyExpiry)
		assert.Equal(t, kept, found, token.Selector)
	}
}

func TestDeletingAUserDeletesItsRememberTokens(t *testing.T) {
	t.Parallel()

	fixture := newRememberTokenFixture(t)
	token := fixture.issue(t, noon().Add(time.Hour))

	deleteUser(t, fixture.pool, fixture.userID)

	_, consumed := fixture.consume(t, token, noon())
	assert.False(t, consumed)
}
