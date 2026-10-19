//go:build integration

package postgres_test

import (
	"bytes"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/aesgcm"
	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testdb"
)

const encryptionKeyFill = 0x01

func openTestDatabase(t *testing.T) *postgres.Pool {
	t.Helper()

	return open(t, testdb.New(t).PostgresConfig())
}

func newUserID() domain.UserID {
	return domain.UserID(uuid.NewV4().String())
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
