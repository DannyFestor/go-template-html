package totp

import (
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec // RFC 6238 and every authenticator app default to HMAC-SHA1, which HOTP still holds.
	"encoding/binary"
	"fmt"
	"time"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

const (
	period            = 30 * time.Second
	digits            = 6
	digitsModulus     = 1_000_000
	dynamicOffsetMask = 0x0f
	signBitMask       = 0x7fffffff
)

func Code(secret domain.TwoFactorSecret, at time.Time) string {
	return codeAt(secret, stepAt(at))
}

func stepAt(at time.Time) domain.TimeStep {
	return domain.TimeStep(at.Unix() / int64(period.Seconds()))
}

func codeAt(secret domain.TwoFactorSecret, step domain.TimeStep) string {
	return fmt.Sprintf("%0*d", digits, rfc4226DynamicTruncation(hmacSHA1(secret, step))%digitsModulus)
}

func hmacSHA1(secret domain.TwoFactorSecret, step domain.TimeStep) []byte {
	var counter [8]byte

	binary.BigEndian.PutUint64(counter[:], uint64(step)) //nolint:gosec // A step counts up from the Unix epoch.

	mac := hmac.New(sha1.New, secret)
	mac.Write(counter[:])

	return mac.Sum(nil)
}

func rfc4226DynamicTruncation(sum []byte) uint32 {
	offset := sum[len(sum)-1] & dynamicOffsetMask

	return binary.BigEndian.Uint32(sum[offset:offset+4]) & signBitMask
}
