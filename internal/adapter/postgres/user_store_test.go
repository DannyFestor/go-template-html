//go:build integration

package postgres_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/adapter/recoverycode"
	"github.com/DannyFestor/go-template-html/internal/adapter/totp"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

const fixedUserID = "019a1f6e-3c4b-7d2e-8f10-1234567890ab"

type fixedIDs struct {
	id string
}

func (f fixedIDs) NewID() string {
	return f.id
}

func TestFindReturnsTheCreatedUser(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)
	store := postgres.NewUserStore(openTestDatabase(t), fixedIDs{id: fixedUserID}, clockAt(createdAt))

	created, err := store.Create(t.Context(), ada("Ada@Example.com"))
	require.NoError(t, err)

	found, ok, err := store.Find(t.Context(), created.ID)
	require.NoError(t, err)
	assert.True(t, ok)

	for name, user := range map[string]domain.User{"created": created, "found": found} {
		assert.WithinDuration(t, createdAt, user.CreatedAt, 0, name)

		user.CreatedAt = createdAt
		assert.Equal(t, domain.User{
			ID:              fixedUserID,
			Name:            "Ada Lovelace",
			Email:           "Ada@Example.com",
			PasswordHash:    storedHash,
			EmailVerifiedAt: time.Time{},
			CreatedAt:       createdAt,
		}, user, name)
	}
}

func TestFindReportsAnUnknownUser(t *testing.T) {
	t.Parallel()

	_, ok, err := newUserStore(openTestDatabase(t)).Find(t.Context(), newUserID())

	require.NoError(t, err)
	assert.False(t, ok)
}

func TestCreateRefusesAnEmailTakenInAnyCase(t *testing.T) {
	t.Parallel()

	store := newUserStore(openTestDatabase(t))
	_, err := store.Create(t.Context(), ada("ada@example.com"))
	require.NoError(t, err)

	_, err = store.Create(t.Context(), ada("ADA@example.com"))

	require.ErrorIs(t, err, domain.ErrEmailTaken)
}

func TestDeletingAUserDeletesItsTwoFactorSecretAndRecoveryCodes(t *testing.T) {
	t.Parallel()

	pool := openTestDatabase(t)
	user := createUser(t, pool)
	secrets := postgres.NewTwoFactorSecretStore(pool, newEncrypter(t, encryptionKeyFill))
	codes := newRecoveryCodeStore(t, pool)
	require.NoError(t, secrets.Save(t.Context(), user, totp.NewSecret()))
	require.NoError(t, codes.Add(t.Context(), user, recoverycode.NewSet()))

	deleteUser(t, pool, user)

	_, found, err := secrets.Find(t.Context(), user)
	require.NoError(t, err)
	assert.False(t, found, "two-factor secret")

	remaining, err := codes.List(t.Context(), user)
	require.NoError(t, err)
	assert.Empty(t, remaining, "recovery codes")
}

func TestFindByEmailFindsTheUserInAnyCase(t *testing.T) {
	t.Parallel()

	store := newUserStore(openTestDatabase(t))
	created, err := store.Create(t.Context(), ada("Ada@Example.com"))
	require.NoError(t, err)

	found, ok, err := store.FindByEmail(t.Context(), "ADA@example.COM")

	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "Ada@Example.com", found.Email)
}

func TestFindByEmailReportsAnUnknownEmail(t *testing.T) {
	t.Parallel()

	_, ok, err := newUserStore(openTestDatabase(t)).FindByEmail(t.Context(), "nobody@example.com")

	require.NoError(t, err)
	assert.False(t, ok)
}

func TestUpdatePasswordHashReplacesTheStoredHash(t *testing.T) {
	t.Parallel()

	const replacement domain.PasswordHash = "$argon2id$v=19$m=16,t=1,p=1$bmV3$a2V5"

	pool := openTestDatabase(t)
	store := newUserStore(pool)
	user := createUser(t, pool)

	require.NoError(t, store.UpdatePasswordHash(t.Context(), user, replacement))

	found, _, err := store.Find(t.Context(), user)
	require.NoError(t, err)
	assert.Equal(t, replacement, found.PasswordHash)
}

func TestAUserCreatedAsVerifiedIsVerifiedFromItsCreation(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC)
	store := postgres.NewUserStore(openTestDatabase(t), fixedIDs{id: fixedUserID}, clockAt(createdAt))
	verified := ada("ada@example.com")
	verified.EmailVerified = true

	created, err := store.Create(t.Context(), verified)
	require.NoError(t, err)

	found, _, err := store.Find(t.Context(), created.ID)
	require.NoError(t, err)

	for name, user := range map[string]domain.User{"created": created, "found": found} {
		assert.True(t, user.Verified(), name)
		assert.WithinDuration(t, createdAt, user.EmailVerifiedAt, 0, name)
	}
}

func TestMarkEmailVerifiedStampsTheFirstVerificationOnly(t *testing.T) {
	t.Parallel()

	firstVerification := time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC)
	pool := openTestDatabase(t)
	store := newUserStore(pool)
	user := createUser(t, pool)

	require.NoError(t, store.MarkEmailVerified(t.Context(), user, firstVerification))
	require.NoError(t, store.MarkEmailVerified(t.Context(), user, firstVerification.Add(time.Hour)))

	found, _, err := store.Find(t.Context(), user)
	require.NoError(t, err)
	assert.True(t, found.Verified())
	assert.WithinDuration(t, firstVerification, found.EmailVerifiedAt, 0)
}

func ada(email string) domain.NewUser {
	return domain.NewUser{Name: "Ada Lovelace", Email: email, PasswordHash: storedHash, EmailVerified: false}
}
