//go:build integration

package postgres_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/aesgcm"
	"github.com/DannyFestor/go-template-html/internal/adapter/hmacdigest"
	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/adapter/recoverycode"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

const digestKeyFill = 0x03

type recoveryCodeFixture struct {
	pool  *postgres.Pool
	store *postgres.RecoveryCodeStore
	user  domain.UserID
	codes []domain.RecoveryCode
}

func TestAUsedRecoveryCodeCannotBeUsedAgain(t *testing.T) {
	t.Parallel()

	fixture := newRecoveryCodeFixture(t)

	assert.True(t, fixture.consume(t, fixture.user, fixture.codes[0]), "first use")
	assert.False(t, fixture.consume(t, fixture.user, fixture.codes[0]), "second use")
}

func TestAnUnknownRecoveryCodeCannotBeUsed(t *testing.T) {
	t.Parallel()

	fixture := newRecoveryCodeFixture(t)

	assert.False(t, fixture.consume(t, fixture.user, recoverycode.NewSet()[0]))
}

func TestARecoveryCodeOfAnotherUserCannotBeUsed(t *testing.T) {
	t.Parallel()

	fixture := newRecoveryCodeFixture(t)

	assert.False(t, fixture.consume(t, newUserID(), fixture.codes[0]))
	assert.True(t, fixture.consume(t, fixture.user, fixture.codes[0]))
}

func TestListReturnsTheUnusedRecoveryCodes(t *testing.T) {
	t.Parallel()

	fixture := newRecoveryCodeFixture(t)

	require.True(t, fixture.consume(t, fixture.user, fixture.codes[0]))

	assert.ElementsMatch(t, fixture.codes[1:], fixture.list(t, fixture.user))
}

func TestDeleteAllRemovesEveryRecoveryCodeOfTheUserOnly(t *testing.T) {
	t.Parallel()

	fixture := newRecoveryCodeFixture(t)
	other := newUserID()
	otherCodes := recoverycode.NewSet()
	require.NoError(t, fixture.store.Add(t.Context(), other, otherCodes))

	require.NoError(t, fixture.store.DeleteAll(t.Context(), fixture.user))

	assert.Empty(t, fixture.list(t, fixture.user))
	assert.ElementsMatch(t, otherCodes, fixture.list(t, other))
}

func TestRecoveryCodesAreStoredEncryptedUnderTheGivenKey(t *testing.T) {
	t.Parallel()

	fixture := newRecoveryCodeFixture(t)
	otherKeyStore := postgres.NewRecoveryCodeStore(
		fixture.pool,
		newEncrypter(t, encryptionKeyFill+1),
		hmacdigest.NewDigester(key(digestKeyFill)),
	)

	_, err := otherKeyStore.List(t.Context(), fixture.user)

	require.ErrorIs(t, err, aesgcm.ErrUndecryptable)
}

func TestRecoveryCodesAreLookedUpByTheirDigestUnderTheGivenKey(t *testing.T) {
	t.Parallel()

	fixture := newRecoveryCodeFixture(t)
	otherKeyStore := postgres.NewRecoveryCodeStore(
		fixture.pool,
		newEncrypter(t, encryptionKeyFill),
		hmacdigest.NewDigester(key(digestKeyFill+1)),
	)

	used, err := otherKeyStore.Consume(t.Context(), fixture.user, fixture.codes[0])

	require.NoError(t, err)
	assert.False(t, used)
}

func newRecoveryCodeFixture(t *testing.T) *recoveryCodeFixture {
	t.Helper()

	pool := openTestDatabase(t)
	fixture := &recoveryCodeFixture{
		pool: pool,
		store: postgres.NewRecoveryCodeStore(
			pool,
			newEncrypter(t, encryptionKeyFill),
			hmacdigest.NewDigester(key(digestKeyFill)),
		),
		user:  newUserID(),
		codes: recoverycode.NewSet(),
	}

	require.NoError(t, fixture.store.Add(t.Context(), fixture.user, fixture.codes))

	return fixture
}

func (f *recoveryCodeFixture) consume(t *testing.T, user domain.UserID, code domain.RecoveryCode) bool {
	t.Helper()

	used, err := f.store.Consume(t.Context(), user, code)
	require.NoError(t, err)

	return used
}

func (f *recoveryCodeFixture) list(t *testing.T, user domain.UserID) []domain.RecoveryCode {
	t.Helper()

	codes, err := f.store.List(t.Context(), user)
	require.NoError(t, err)

	return codes
}
