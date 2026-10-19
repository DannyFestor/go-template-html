//go:build integration

package postgres_test

import (
	"bytes"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/aesgcm"
	"github.com/DannyFestor/go-template-html/internal/adapter/clock"
	"github.com/DannyFestor/go-template-html/internal/adapter/hmacdigest"
	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/adapter/uuidv7"
	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testdb"
)

const (
	encryptionKeyFill                     = 0x01
	storedHash        domain.PasswordHash = "$argon2id$v=19$m=8,t=1,p=1$c2FsdA$a2V5"
)

func openTestDatabase(t *testing.T) *postgres.Pool {
	t.Helper()

	return open(t, testdb.New(t).PostgresConfig())
}

func newUserID() domain.UserID {
	return domain.UserID(uuid.NewV4().String())
}

func newUserStore(pool *postgres.Pool) *postgres.UserStore {
	return postgres.NewUserStore(pool, uuidv7.NewGenerator(clock.System{}), clock.System{})
}

func createUser(t *testing.T, pool *postgres.Pool) domain.UserID {
	t.Helper()

	user, err := newUserStore(pool).Create(t.Context(), domain.NewUser{
		Name:         "Grace Hopper",
		Email:        uuid.NewV4().String() + "@example.com",
		PasswordHash: storedHash,
	})
	require.NoError(t, err)

	return user.ID
}

func newRecoveryCodeStore(t *testing.T, pool *postgres.Pool) *postgres.RecoveryCodeStore {
	t.Helper()

	return postgres.NewRecoveryCodeStore(
		pool,
		newEncrypter(t, encryptionKeyFill),
		hmacdigest.NewDigester(key(digestKeyFill)),
	)
}

func newEncrypter(t *testing.T, keyFill byte) *aesgcm.Encrypter {
	t.Helper()

	encrypter, err := aesgcm.NewEncrypter(key(keyFill))
	require.NoError(t, err)

	return encrypter
}

func key(fill byte) []byte {
	return bytes.Repeat([]byte{fill}, 32)
}
