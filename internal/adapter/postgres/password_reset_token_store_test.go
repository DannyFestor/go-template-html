//go:build integration

package postgres_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/adapter/resettoken"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

type passwordResetTokenFixture struct {
	pool   *postgres.Pool
	store  *postgres.PasswordResetTokenStore
	userID domain.UserID
}

func newPasswordResetTokenFixture(t *testing.T) passwordResetTokenFixture {
	t.Helper()

	pool := openTestDatabase(t)

	return passwordResetTokenFixture{
		pool:   pool,
		store:  postgres.NewPasswordResetTokenStore(pool),
		userID: createUser(t, pool),
	}
}

func (f passwordResetTokenFixture) replace(
	t *testing.T,
	userID domain.UserID,
	expiresAt time.Time,
) domain.PasswordResetToken {
	t.Helper()

	token := resettoken.Generator{}.New()
	require.NoError(t, f.store.Replace(t.Context(), userID, token, expiresAt))

	return token
}

func (f passwordResetTokenFixture) consume(
	t *testing.T,
	token domain.PasswordResetToken,
	now time.Time,
) (domain.UserID, bool) {
	t.Helper()

	userID, consumed, err := f.store.Consume(t.Context(), token, now)
	require.NoError(t, err)

	return userID, consumed
}

func TestAPasswordResetTokenIsConsumedForItsUser(t *testing.T) {
	t.Parallel()

	fixture := newPasswordResetTokenFixture(t)
	token := fixture.replace(t, fixture.userID, noon().Add(time.Hour))

	userID, consumed := fixture.consume(t, token, noon())

	assert.True(t, consumed)
	assert.Equal(t, fixture.userID, userID)
}

func TestAPasswordResetTokenIsConsumedOnlyOnce(t *testing.T) {
	t.Parallel()

	fixture := newPasswordResetTokenFixture(t)
	token := fixture.replace(t, fixture.userID, noon().Add(time.Hour))
	fixture.consume(t, token, noon())

	_, consumedAgain := fixture.consume(t, token, noon())

	assert.False(t, consumedAgain)
}

func TestAPasswordResetTokenIsNotConsumedOnceItExpires(t *testing.T) {
	t.Parallel()

	fixture := newPasswordResetTokenFixture(t)
	token := fixture.replace(t, fixture.userID, noon())

	_, consumed := fixture.consume(t, token, noon())

	assert.False(t, consumed)
}

func TestAnUnknownPasswordResetTokenIsNotConsumed(t *testing.T) {
	t.Parallel()

	fixture := newPasswordResetTokenFixture(t)
	fixture.replace(t, fixture.userID, noon().Add(time.Hour))

	_, consumed := fixture.consume(t, resettoken.Generator{}.New(), noon())

	assert.False(t, consumed)
}

func TestANewPasswordResetTokenReplacesTheUsersPreviousOne(t *testing.T) {
	t.Parallel()

	fixture := newPasswordResetTokenFixture(t)
	previous := fixture.replace(t, fixture.userID, noon().Add(time.Hour))
	current := fixture.replace(t, fixture.userID, noon().Add(time.Hour))

	_, previousConsumed := fixture.consume(t, previous, noon())
	_, currentConsumed := fixture.consume(t, current, noon())

	assert.False(t, previousConsumed)
	assert.True(t, currentConsumed)
}

func TestAPasswordResetTokenLeavesOtherUsersTokensAlone(t *testing.T) {
	t.Parallel()

	fixture := newPasswordResetTokenFixture(t)
	other := createUser(t, fixture.pool)
	othersToken := fixture.replace(t, other, noon().Add(time.Hour))
	fixture.replace(t, fixture.userID, noon().Add(time.Hour))

	userID, consumed := fixture.consume(t, othersToken, noon())

	assert.True(t, consumed)
	assert.Equal(t, other, userID)
}

func TestDeletingAUserDeletesItsPasswordResetToken(t *testing.T) {
	t.Parallel()

	fixture := newPasswordResetTokenFixture(t)
	token := fixture.replace(t, fixture.userID, noon().Add(time.Hour))

	deleteUser(t, fixture.pool, fixture.userID)

	_, consumed := fixture.consume(t, token, noon())
	assert.False(t, consumed)
}
