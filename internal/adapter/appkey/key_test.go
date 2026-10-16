package appkey_test

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/appkey"
)

const encodedKey = "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8="

func TestParseAcceptsThirtyTwoBytesOfStandardBase64(t *testing.T) {
	t.Parallel()

	_, err := appkey.Parse(encodedKey)

	assert.NoError(t, err)
}

func TestParseRejectsAMalformedKeyWithoutEchoingIt(t *testing.T) {
	t.Parallel()

	for name, encoded := range map[string]string{
		"empty":      "",
		"not base64": "not-a-key!",
		"too short":  base64.StdEncoding.EncodeToString(make([]byte, appkey.Size-1)),
		"too long":   base64.StdEncoding.EncodeToString(make([]byte, appkey.Size+1)),
		"url base64": base64.URLEncoding.EncodeToString(bytes.Repeat([]byte{0xfb}, appkey.Size)),
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := appkey.Parse(encoded)

			require.ErrorIs(t, err, appkey.ErrMalformedKey)

			if encoded != "" {
				assert.NotContains(t, err.Error(), encoded)
			}
		})
	}
}

func TestSubkeysDifferPerPurpose(t *testing.T) {
	t.Parallel()

	key := mustParse(t, encodedKey)
	seen := map[string]appkey.Purpose{}

	for _, purpose := range appkey.Purposes() {
		subkey := mustSubkey(t, key, purpose)

		require.Len(t, subkey, appkey.SubkeySize)
		assert.NotContains(t, seen, string(subkey), "%s and %s share a subkey", purpose, seen[string(subkey)])

		seen[string(subkey)] = purpose
	}

	assert.Len(t, seen, 5)
}

func TestASubkeyIsStableForItsKeyAndPurpose(t *testing.T) {
	t.Parallel()

	first := mustSubkey(t, mustParse(t, encodedKey), appkey.PurposeSignedURL)
	second := mustSubkey(t, mustParse(t, encodedKey), appkey.PurposeSignedURL)

	assert.Equal(t, first, second)
}

func TestASubkeyIsNeitherTheAppKeyNorSharedWithAnotherAppKey(t *testing.T) {
	t.Parallel()

	raw, err := base64.StdEncoding.DecodeString(encodedKey)
	require.NoError(t, err)

	otherKey := mustParse(t, base64.StdEncoding.EncodeToString(make([]byte, appkey.Size)))
	subkey := mustSubkey(t, mustParse(t, encodedKey), appkey.PurposeSignedURL)

	assert.NotEqual(t, raw, subkey)
	assert.NotEqual(t, mustSubkey(t, otherKey, appkey.PurposeSignedURL), subkey)
}

func mustParse(t *testing.T, encoded string) appkey.Key {
	t.Helper()

	key, err := appkey.Parse(encoded)
	require.NoError(t, err)

	return key
}

func mustSubkey(t *testing.T, key appkey.Key, purpose appkey.Purpose) []byte {
	t.Helper()

	subkey, err := key.Subkey(purpose)
	require.NoError(t, err)

	return subkey
}
