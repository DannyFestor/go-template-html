package fingerprint_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/adapter/fingerprint"
	"github.com/DannyFestor/go-template-html/internal/adapter/hmacdigest"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

const (
	hash      domain.PasswordHash = "$argon2id$v=19$m=8,t=1,p=1$c2FsdA$a2V5"
	otherHash domain.PasswordHash = "$argon2id$v=19$m=8,t=1,p=1$c2FsdA$b3RoZXI"
)

func newCredentials(keyFill byte) *fingerprint.Credentials {
	return fingerprint.NewCredentials(hmacdigest.NewDigester(bytes.Repeat([]byte{keyFill}, 32)))
}

func TestAFingerprintMatchesTheHashItWasTakenOf(t *testing.T) {
	t.Parallel()

	credentials := newCredentials(0x01)

	assert.True(t, credentials.Matches(hash, credentials.Of(hash)))
}

func TestAFingerprintDoesNotMatchAnotherHash(t *testing.T) {
	t.Parallel()

	credentials := newCredentials(0x01)

	assert.False(t, credentials.Matches(otherHash, credentials.Of(hash)))
}

func TestAFingerprintDoesNotMatchUnderAnotherKey(t *testing.T) {
	t.Parallel()

	assert.False(t, newCredentials(0x02).Matches(hash, newCredentials(0x01).Of(hash)))
}

func TestAFingerprintDoesNotRevealTheHash(t *testing.T) {
	t.Parallel()

	assert.NotContains(t, newCredentials(0x01).Of(hash), "a2V5")
}
