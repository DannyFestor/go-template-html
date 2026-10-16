package totp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/adapter/totp"
)

func TestNewSecretHas160RandomBits(t *testing.T) {
	t.Parallel()

	first, second := totp.NewSecret(), totp.NewSecret()

	assert.Len(t, first, 20)
	assert.NotEqual(t, first, second)
}

func TestURIEncodesTheSecretInBase32ForAuthenticatorApps(t *testing.T) {
	t.Parallel()

	uri := totp.URI("Go Template & Co", "alice@example.com", seed())

	assert.Equal(t,
		"otpauth://totp/Go%20Template%20&%20Co:alice@example.com"+
			"?secret=GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ&issuer=Go%20Template%20%26%20Co",
		uri,
	)
}
