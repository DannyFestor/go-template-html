package hmacdigest_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/adapter/hmacdigest"
)

const message = "ABCDEFGHIJ-KLMNOPQRST"

func TestDigestIsTheHMACSHA256OfTheMessage(t *testing.T) {
	t.Parallel()

	digest := hmacdigest.NewDigester(key(0x01)).Digest([]byte(message))

	assert.Equal(t, "c8939ec4e281bd823043a053dd9883ec64ab6d5ec301f9c0efd823e41892ba87", hex.EncodeToString(digest))
}

func TestDigestDiffersPerKey(t *testing.T) {
	t.Parallel()

	first := hmacdigest.NewDigester(key(0x01)).Digest([]byte(message))
	second := hmacdigest.NewDigester(key(0x02)).Digest([]byte(message))

	assert.NotEqual(t, first, second)
}

func key(fill byte) []byte {
	return bytes.Repeat([]byte{fill}, 32)
}
