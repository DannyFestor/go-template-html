package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
)

const adasLinkHash = "hash-of-ada@example.com"

var errUsersDown = errors.New("users down")

type verifyEmailDoubles struct {
	links *auth.MockVerificationLinkMatcher
	users *auth.MockEmailVerifier
}

func newVerifyEmailDoubles(t *testing.T) verifyEmailDoubles {
	t.Helper()

	links := auth.NewMockVerificationLinkMatcher(gomock.NewController(t))
	links.EXPECT().
		Matches(gomock.Any(), gomock.Any()).
		DoAndReturn(func(email, hash string) bool { return hash == "hash-of-"+email }).
		AnyTimes()

	return verifyEmailDoubles{links: links, users: auth.NewMockEmailVerifier(gomock.NewController(t))}
}

func (d verifyEmailDoubles) verifyEmail(t *testing.T) *auth.VerifyEmail {
	t.Helper()

	return auth.NewVerifyEmail(d.links, d.users, fixedClock(t))
}

func adasLink() auth.VerifyEmailIn {
	return auth.VerifyEmailIn{User: storedAda(), LinkUserID: storedAda().ID, LinkHash: adasLinkHash}
}

func TestVerifyEmailMarksTheEmailVerifiedNow(t *testing.T) {
	t.Parallel()

	doubles := newVerifyEmailDoubles(t)
	doubles.users.EXPECT().
		MarkEmailVerified(gomock.Any(), storedAda().ID, time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC))

	_, err := doubles.verifyEmail(t).Run(t.Context(), adasLink())

	require.NoError(t, err)
}

func TestVerifyEmailRefusesALinkIssuedForAnotherUserOrAddress(t *testing.T) {
	t.Parallel()

	for name, link := range map[string]func(*auth.VerifyEmailIn){
		"another user":    func(in *auth.VerifyEmailIn) { in.LinkUserID = "019a1f6e-3c4b-7d2e-8f10-0000000000ff" },
		"another address": func(in *auth.VerifyEmailIn) { in.LinkHash = "hash-of-grace@example.com" },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			in := adasLink()
			link(&in)

			_, err := newVerifyEmailDoubles(t).verifyEmail(t).Run(t.Context(), in)

			require.ErrorIs(t, err, auth.ErrVerificationLinkMismatch)
		})
	}
}

func TestVerifyEmailFailsWhenTheVerificationCannotBeStored(t *testing.T) {
	t.Parallel()

	doubles := newVerifyEmailDoubles(t)
	doubles.users.EXPECT().MarkEmailVerified(gomock.Any(), gomock.Any(), gomock.Any()).Return(errUsersDown)

	_, err := doubles.verifyEmail(t).Run(t.Context(), adasLink())

	require.ErrorIs(t, err, errUsersDown)
}
