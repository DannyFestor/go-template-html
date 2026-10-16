package signedurl_test

import (
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/signedurl"
)

func TestVerifyAcceptsASignedURL(t *testing.T) {
	t.Parallel()

	signed := newSigner(t, keyFill, signingTime()).Sign(path, url.Values{"hash": {"abc"}, "id": {"42"}})

	assert.NoError(t, newVerifier(t, signingTime().Add(365*24*time.Hour)).Verify(parse(t, signed)))
}

func TestVerifyAcceptsTheQueryInAnyOrder(t *testing.T) {
	t.Parallel()

	signed := parse(t, newSigner(t, keyFill, signingTime()).Sign(path, url.Values{"a": {"1"}, "b": {"2"}}))
	signed.RawQuery = "signature=" + signed.Query().Get("signature") + "&b=2&a=1"

	assert.NoError(t, newVerifier(t, signingTime()).Verify(signed))
}

func TestVerifyRejectsATamperedURL(t *testing.T) {
	t.Parallel()

	signed := newSigner(t, keyFill, signingTime()).SignTemporary(path, url.Values{"hash": {"abc"}})

	tamperings := map[string]func(*url.URL){
		"path":                func(u *url.URL) { u.Path = "/email/verify/43" },
		"query value":         func(u *url.URL) { setQuery(u, "hash", "abd") },
		"added parameter":     func(u *url.URL) { setQuery(u, "admin", "1") },
		"removed parameter":   func(u *url.URL) { deleteQuery(u, "hash") },
		"extended expiry":     func(u *url.URL) { setQuery(u, "expires", "1767229200") },
		"removed expiry":      func(u *url.URL) { deleteQuery(u, "expires") },
		"removed signature":   func(u *url.URL) { deleteQuery(u, "signature") },
		"empty signature":     func(u *url.URL) { setQuery(u, "signature", "") },
		"non-hex signature":   func(u *url.URL) { setQuery(u, "signature", strings.Repeat("z", 64)) },
		"truncated signature": func(u *url.URL) { setQuery(u, "signature", u.Query().Get("signature")[:62]) },
		"flipped signature":   func(u *url.URL) { setQuery(u, "signature", flipLastHexDigit(u.Query().Get("signature"))) },
	}

	for name, tamper := range tamperings {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tampered := parse(t, signed)
			tamper(tampered)

			err := newVerifier(t, signingTime()).Verify(tampered)

			assert.ErrorIs(t, err, signedurl.ErrInvalidSignature)
		})
	}
}

func TestVerifyRejectsAURLSignedWithAnotherKey(t *testing.T) {
	t.Parallel()

	signed := newSigner(t, keyFill+1, signingTime()).Sign(path, url.Values{})

	err := newVerifier(t, signingTime()).Verify(parse(t, signed))

	assert.ErrorIs(t, err, signedurl.ErrInvalidSignature)
}

func TestVerifyAcceptsATemporaryURLUntilItExpires(t *testing.T) {
	t.Parallel()

	signed := parse(t, newSigner(t, keyFill, signingTime()).SignTemporary(path, url.Values{}))

	assert.NoError(t, newVerifier(t, signingTime().Add(ttl-time.Second)).Verify(signed))
}

func TestVerifyRejectsAnExpiredTemporaryURL(t *testing.T) {
	t.Parallel()

	signed := parse(t, newSigner(t, keyFill, signingTime()).SignTemporaryFor(path, url.Values{}, time.Minute))

	for _, elapsed := range []time.Duration{time.Minute, time.Hour} {
		err := newVerifier(t, signingTime().Add(elapsed)).Verify(signed)

		require.ErrorIs(t, err, signedurl.ErrExpired, "after %s", elapsed)
	}
}

func newVerifier(t *testing.T, now time.Time) *signedurl.Verifier {
	t.Helper()

	return signedurl.NewVerifier(key(keyFill), clockAt(t, now))
}

func setQuery(target *url.URL, name, value string) {
	query := target.Query()
	query.Set(name, value)
	target.RawQuery = query.Encode()
}

func deleteQuery(target *url.URL, name string) {
	query := target.Query()
	query.Del(name)
	target.RawQuery = query.Encode()
}

func flipLastHexDigit(signature string) string {
	last := "0"
	if strings.HasSuffix(signature, "0") {
		last = "1"
	}

	return signature[:len(signature)-1] + last
}
