package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

const (
	passwordResetTTL                           = time.Hour
	resetToken       domain.PasswordResetToken = "RESETTOKEN"
	resetURL                                   = "https://example.com/reset-password?token=RESETTOKEN"
)

type issuePasswordResetLinkDoubles struct {
	generator *auth.MockPasswordResetTokenGenerator
	tokens    *auth.MockPasswordResetTokenReplacer
	links     *auth.MockPasswordResetLinker
}

func newIssuePasswordResetLinkDoubles(t *testing.T) issuePasswordResetLinkDoubles {
	t.Helper()

	generator := auth.NewMockPasswordResetTokenGenerator(gomock.NewController(t))
	generator.EXPECT().New().Return(resetToken)

	return issuePasswordResetLinkDoubles{
		generator: generator,
		tokens:    auth.NewMockPasswordResetTokenReplacer(gomock.NewController(t)),
		links:     auth.NewMockPasswordResetLinker(gomock.NewController(t)),
	}
}

func (d issuePasswordResetLinkDoubles) issue(t *testing.T) *auth.IssuePasswordResetLink {
	t.Helper()

	return auth.NewIssuePasswordResetLink(d.generator, d.tokens, d.links, fixedClock(t), passwordResetTTL)
}

func TestIssuePasswordResetLinkLinksToAFreshTokenThatReplacesTheUsersOldOne(t *testing.T) {
	t.Parallel()

	doubles := newIssuePasswordResetLinkDoubles(t)
	doubles.tokens.EXPECT().
		Replace(gomock.Any(), storedAda().ID, resetToken, time.Date(2026, 10, 21, 10, 0, 0, 0, time.UTC)).
		Return(nil)
	doubles.links.EXPECT().PasswordResetURL(resetToken).Return(resetURL)

	issued, err := doubles.issue(t).Run(t.Context(), auth.IssuePasswordResetLinkIn{UserID: storedAda().ID})

	require.NoError(t, err)
	assert.Equal(t, auth.IssuePasswordResetLinkOut{URL: resetURL}, issued)
}

func TestIssuePasswordResetLinkFailsWhenTheTokenCannotBeStored(t *testing.T) {
	t.Parallel()

	doubles := newIssuePasswordResetLinkDoubles(t)
	doubles.tokens.EXPECT().Replace(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errStore)

	_, err := doubles.issue(t).Run(t.Context(), auth.IssuePasswordResetLinkIn{UserID: storedAda().ID})

	require.ErrorIs(t, err, errStore)
}
