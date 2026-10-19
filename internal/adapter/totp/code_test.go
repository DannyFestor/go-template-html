package totp_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/adapter/totp"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

const rfc6238SHA1Seed = "12345678901234567890"

func TestCodeIsTheLastSixDigitsOfTheRFC6238SHA1Vectors(t *testing.T) {
	t.Parallel()

	vectors := map[int64]string{
		59:          "287082",
		1111111109:  "081804",
		1111111111:  "050471",
		1234567890:  "005924",
		2000000000:  "279037",
		20000000000: "353130",
	}

	for unix, code := range vectors {
		assert.Equal(t, code, totp.Code(domain.TwoFactorSecret(rfc6238SHA1Seed), time.Unix(unix, 0)), "at %d", unix)
	}
}
