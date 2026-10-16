package aesgcm_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/aesgcm"
)

const plaintext = "JBSWY3DPEHPK3PXP"

func TestDecryptReturnsWhatEncryptSealed(t *testing.T) {
	t.Parallel()

	encrypter := newEncrypter(t, 1)

	opened, err := encrypter.Decrypt(encrypter.Encrypt([]byte(plaintext)))

	require.NoError(t, err)
	assert.Equal(t, plaintext, string(opened))
}

func TestEncryptingTheSamePlaintextTwiceGivesDifferentCiphertexts(t *testing.T) {
	t.Parallel()

	encrypter := newEncrypter(t, 1)

	first, second := encrypter.Encrypt([]byte(plaintext)), encrypter.Encrypt([]byte(plaintext))

	assert.NotEqual(t, first, second)
	assert.NotContains(t, string(first), plaintext)
}

func TestDecryptRejectsATamperedCiphertext(t *testing.T) {
	t.Parallel()

	encrypter := newEncrypter(t, 1)
	sealed := encrypter.Encrypt([]byte(plaintext))

	for index := range sealed {
		tampered := bytes.Clone(sealed)
		tampered[index] ^= 0x01

		_, err := encrypter.Decrypt(tampered)

		require.ErrorIs(t, err, aesgcm.ErrUndecryptable, "byte %d flipped", index)
	}
}

func TestDecryptRejectsATruncatedCiphertext(t *testing.T) {
	t.Parallel()

	encrypter := newEncrypter(t, 1)
	sealed := encrypter.Encrypt([]byte(plaintext))

	for _, length := range []int{0, 1, len(sealed) - 1} {
		_, err := encrypter.Decrypt(sealed[:length])

		require.ErrorIs(t, err, aesgcm.ErrUndecryptable, "%d bytes", length)
	}
}

func TestDecryptRejectsACiphertextSealedUnderAnotherKey(t *testing.T) {
	t.Parallel()

	sealed := newEncrypter(t, 1).Encrypt([]byte(plaintext))

	_, err := newEncrypter(t, 2).Decrypt(sealed)

	assert.ErrorIs(t, err, aesgcm.ErrUndecryptable)
}

func TestNewEncrypterRejectsAKeyThatIsNotAES256(t *testing.T) {
	t.Parallel()

	for _, length := range []int{0, 16, 24, 31, 33} {
		_, err := aesgcm.NewEncrypter(make([]byte, length))

		assert.Error(t, err, "%d bytes", length)
	}
}

func newEncrypter(t *testing.T, fill byte) *aesgcm.Encrypter {
	t.Helper()

	encrypter, err := aesgcm.NewEncrypter(bytes.Repeat([]byte{fill}, 32))
	require.NoError(t, err)

	return encrypter
}
