//go:build integration

package postgres_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/adapter/remembertoken"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

type rememberTokenFixture struct {
	pool   *postgres.Pool
	store  *postgres.RememberTokenStore
	userID domain.UserID
}

func newRememberTokenFixture(t *testing.T) rememberTokenFixture {
	t.Helper()

	pool := openTestDatabase(t)

	return rememberTokenFixture{pool: pool, store: postgres.NewRememberTokenStore(pool), userID: createUser(t, pool)}
}

func (f rememberTokenFixture) issue(t *testing.T, expiresAt time.Time) domain.RememberToken {
	t.Helper()

	token := remembertoken.Generator{}.New()
	require.NoError(t, f.store.Issue(t.Context(), f.userID, token, expiresAt))

	return token
}

func (f rememberTokenFixture) consume(t *testing.T, token domain.RememberToken, now time.Time) (domain.UserID, bool) {
	t.Helper()

	userID, consumed, err := f.store.Consume(t.Context(), token, now)
	require.NoError(t, err)

	return userID, consumed
}

func noon() time.Time {
	return time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
}

func TestAnIssuedRememberTokenIsConsumedOnceForItsUser(t *testing.T) {
	t.Parallel()

	fixture := newRememberTokenFixture(t)
	token := fixture.issue(t, noon().Add(time.Hour))

	userID, consumed := fixture.consume(t, token, noon())
	assert.True(t, consumed)
	assert.Equal(t, fixture.userID, userID)

	_, consumedAgain := fixture.consume(t, token, noon())
	assert.False(t, consumedAgain)
}

func TestARememberTokenWithAWrongValidatorIsNotConsumed(t *testing.T) {
	t.Parallel()

	fixture := newRememberTokenFixture(t)
	token := fixture.issue(t, noon().Add(time.Hour))
	forged := domain.RememberToken{Selector: token.Selector, Validator: token.Validator + "X"}

	_, consumed := fixture.consume(t, forged, noon())
	assert.False(t, consumed)

	_, consumedWithTheRightValidator := fixture.consume(t, token, noon())
	assert.True(t, consumedWithTheRightValidator)
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

func TestDeleteExpiredDeletesOnlyTheRememberTokensExpiredByNow(t *testing.T) {
	t.Parallel()

	fixture := newRememberTokenFixture(t)
	expired := fixture.issue(t, noon())
	live := fixture.issue(t, noon().Add(time.Second))

	require.NoError(t, fixture.store.DeleteExpired(t.Context(), noon()))

	var beforeEitherExpired time.Time

	_, expiredKept := fixture.consume(t, expired, beforeEitherExpired)
	_, liveKept := fixture.consume(t, live, beforeEitherExpired)
	assert.False(t, expiredKept)
	assert.True(t, liveKept)
}

func TestDeletingAUserDeletesItsRememberTokens(t *testing.T) {
	t.Parallel()

	fixture := newRememberTokenFixture(t)
	token := fixture.issue(t, noon().Add(time.Hour))

	deleteUser(t, fixture.pool, fixture.userID)

	_, consumed := fixture.consume(t, token, noon())
	assert.False(t, consumed)
}
