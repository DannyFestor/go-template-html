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
			ID:           fixedUserID,
			Name:         "Ada Lovelace",
			Email:        "Ada@Example.com",
			PasswordHash: storedHash,
			CreatedAt:    createdAt,
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

	require.NoError(t, postgres.Exec(t.Context(), pool, "DELETE FROM users WHERE id = $1", string(user)))

	_, found, err := secrets.Find(t.Context(), user)
	require.NoError(t, err)
	assert.False(t, found, "two-factor secret")

	remaining, err := codes.List(t.Context(), user)
	require.NoError(t, err)
	assert.Empty(t, remaining, "recovery codes")
}

func ada(email string) domain.NewUser {
	return domain.NewUser{Name: "Ada Lovelace", Email: email, PasswordHash: storedHash}
}
