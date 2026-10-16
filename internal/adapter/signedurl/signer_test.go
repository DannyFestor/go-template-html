package signedurl_test

import (
	"bytes"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/adapter/signedurl"
)

const (
	path    = "/email/verify/42"
	ttl     = 2 * time.Hour
	keyFill = 0x01
)

func TestSignAppendsTheHMACOfThePathAndSortedQuery(t *testing.T) {
	t.Parallel()

	signed := newSigner(t, keyFill, signingTime()).Sign(path, url.Values{"hash": {"abc"}})

	assert.Equal(t,
		"https://example.com/email/verify/42?hash=abc"+
			"&signature=ff6220713d13f59a2a113e14851275507497ee497b3097220fed72cebcedf397",
		signed,
	)
}

func TestSignTemporaryExpiresAfterTheConfiguredTTL(t *testing.T) {
	t.Parallel()

	signed := newSigner(t, keyFill, signingTime()).SignTemporary(path, url.Values{"hash": {"abc"}})

	assert.Equal(t,
		"https://example.com/email/verify/42?expires=1767225600&hash=abc"+
			"&signature=e8c360fab7325455b90f044b87bfec7e6109fa242e78da894940d85609831bcc",
		signed,
	)
}

func TestSignTemporaryForExpiresAfterTheGivenTTL(t *testing.T) {
	t.Parallel()

	signed := newSigner(t, keyFill, signingTime()).SignTemporaryFor(path, url.Values{}, time.Minute)

	assert.Equal(t, "1767218460", parse(t, signed).Query().Get("expires"))
}

func TestSignLeavesTheCallersQueryUntouched(t *testing.T) {
	t.Parallel()

	query := url.Values{"hash": {"abc"}}

	newSigner(t, keyFill, signingTime()).SignTemporary(path, query)

	assert.Equal(t, url.Values{"hash": {"abc"}}, query)
}

func TestSignBuildsOnTheBaseURLsPath(t *testing.T) {
	t.Parallel()

	base := parse(t, "https://example.com/app")
	signer := signedurl.NewSigner(*base, key(keyFill), clockAt(t, signingTime()), ttl)

	signed := signer.Sign(path, url.Values{})

	assert.Equal(t, "https://example.com/app/email/verify/42", stripQuery(t, signed))
	assert.NoError(t, signedurl.NewVerifier(key(keyFill), clockAt(t, signingTime())).Verify(parse(t, signed)))
}

func TestSignSignsThePathItEmits(t *testing.T) {
	t.Parallel()

	signed := newSigner(t, keyFill, signingTime()).Sign("email//verify/../verify/42", url.Values{})

	assert.Equal(t, "https://example.com/email/verify/42", stripQuery(t, signed))
	assert.NoError(t, signedurl.NewVerifier(key(keyFill), clockAt(t, signingTime())).Verify(parse(t, signed)))
}

func signingTime() time.Time {
	return time.Date(2025, time.December, 31, 22, 0, 0, 0, time.UTC)
}

func newSigner(t *testing.T, fill byte, now time.Time) *signedurl.Signer {
	t.Helper()

	return signedurl.NewSigner(*parse(t, "https://example.com"), key(fill), clockAt(t, now), ttl)
}

func key(fill byte) []byte {
	return bytes.Repeat([]byte{fill}, 32)
}

func clockAt(t *testing.T, now time.Time) *signedurl.MockClock {
	t.Helper()

	clock := signedurl.NewMockClock(gomock.NewController(t))
	clock.EXPECT().Now().Return(now).AnyTimes()

	return clock
}

func parse(t *testing.T, raw string) *url.URL {
	t.Helper()

	parsed, err := url.Parse(raw)
	require.NoError(t, err)

	return parsed
}

func stripQuery(t *testing.T, raw string) string {
	t.Helper()

	parsed := parse(t, raw)
	parsed.RawQuery = ""

	return parsed.String()
}
