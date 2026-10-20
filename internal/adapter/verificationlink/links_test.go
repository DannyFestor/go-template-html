package verificationlink_test

import (
	"bytes"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/signedurl"
	"github.com/DannyFestor/go-template-html/internal/adapter/verificationlink"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

const (
	adaID            = "019a1f6e-3c4b-7d2e-8f10-1234567890ab"
	sha256OfAdaEmail = "b5fc85e55755f9e0d030a10ab4429b6b2944855f9a0d60077fe832becbc41d72"
	ttl              = time.Hour
)

type fixedClock struct {
	now time.Time
}

func (c fixedClock) Now() time.Time {
	return c.now
}

func issuedAt() time.Time {
	return time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC)
}

func signingKey() []byte {
	return bytes.Repeat([]byte{0x01}, 32)
}

func newLinks(t *testing.T) *verificationlink.Links {
	t.Helper()

	base, err := url.Parse("https://example.com")
	require.NoError(t, err)

	signer := signedurl.NewSigner(*base, signingKey(), fixedClock{now: issuedAt()}, 2*time.Hour)

	return verificationlink.NewLinks(signer, "/verify-email", ttl)
}

func ada() domain.User {
	return domain.User{
		ID:              adaID,
		Name:            "Ada Lovelace",
		Email:           "ada@example.com",
		PasswordHash:    "",
		EmailVerifiedAt: time.Time{},
		CreatedAt:       issuedAt(),
	}
}

func issuedLink(t *testing.T) *url.URL {
	t.Helper()

	link, err := url.Parse(newLinks(t).For(ada()))
	require.NoError(t, err)

	return link
}

func TestALinkPointsAtTheUserAndTheHashOfTheirEmail(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "/verify-email/"+adaID+"/"+sha256OfAdaEmail, issuedLink(t).Path)
}

func TestALinkIsSignedAndExpiresAfterTheVerificationTTL(t *testing.T) {
	t.Parallel()

	link := issuedLink(t)

	assert.Equal(t, "1792576800", link.Query().Get("expires"))

	for elapsed, want := range map[time.Duration]error{ttl - time.Second: nil, ttl: signedurl.ErrExpired} {
		verifier := signedurl.NewVerifier(signingKey(), fixedClock{now: issuedAt().Add(elapsed)})

		assert.ErrorIs(t, verifier.Verify(link), want, "after %s", elapsed)
	}
}

func TestTheHashMatchesOnlyTheEmailTheLinkWasIssuedFor(t *testing.T) {
	t.Parallel()

	links := newLinks(t)

	assert.True(t, links.Matches("ada@example.com", sha256OfAdaEmail))
	assert.False(t, links.Matches("grace@example.com", sha256OfAdaEmail))
	assert.False(t, links.Matches("ada@example.com", ""))
}
