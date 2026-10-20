package auth_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
)

var errUnreadableHash = errors.New("unreadable hash")

func confirmPasswordVerifying(t *testing.T, password string, matches bool, verifyErr error) *auth.ConfirmPassword {
	t.Helper()

	passwords := auth.NewMockPasswordVerifier(gomock.NewController(t))
	passwords.EXPECT().Verify(password, string(storedHash)).Return(matches, verifyErr)

	return auth.NewConfirmPassword(passwords)
}

func adaConfirms(password string) auth.ConfirmPasswordIn {
	return auth.ConfirmPasswordIn{User: storedAda(), Password: password}
}

func TestConfirmPasswordAcceptsTheUsersCurrentPassword(t *testing.T) {
	t.Parallel()

	_, err := confirmPasswordVerifying(t, rightSecret, true, nil).Run(t.Context(), adaConfirms(rightSecret))

	require.NoError(t, err)
}

func TestConfirmPasswordRefusesAnyOtherPassword(t *testing.T) {
	t.Parallel()

	_, err := confirmPasswordVerifying(t, wrongSecret, false, nil).Run(t.Context(), adaConfirms(wrongSecret))

	require.ErrorIs(t, err, auth.ErrIncorrectPassword)
}

func TestConfirmPasswordFailsWhenTheStoredHashCannotBeVerified(t *testing.T) {
	t.Parallel()

	_, err := confirmPasswordVerifying(t, rightSecret, false, errUnreadableHash).
		Run(t.Context(), adaConfirms(rightSecret))

	require.ErrorIs(t, err, errUnreadableHash)
	assert.NotErrorIs(t, err, auth.ErrIncorrectPassword)
}
