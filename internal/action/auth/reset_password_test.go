package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

type resetPasswordDoubles struct {
	tokens   *auth.MockPasswordResetTokenConsumer
	password *auth.MockPasswordChanger
}

func newResetPasswordDoubles(t *testing.T) resetPasswordDoubles {
	t.Helper()

	return resetPasswordDoubles{
		tokens:   auth.NewMockPasswordResetTokenConsumer(gomock.NewController(t)),
		password: auth.NewMockPasswordChanger(gomock.NewController(t)),
	}
}

func (d resetPasswordDoubles) resetPassword(t *testing.T) *auth.ResetPassword {
	t.Helper()

	return auth.NewResetPassword(inTransaction(t), d.tokens, d.password, fixedClock(t))
}

func resetWith(token domain.PasswordResetToken) auth.ResetPasswordIn {
	return auth.ResetPasswordIn{Token: token, Password: "new password"}
}

func TestResetPasswordConsumesTheTokenAndChangesItsUsersPassword(t *testing.T) {
	t.Parallel()

	doubles := newResetPasswordDoubles(t)
	doubles.tokens.EXPECT().
		Consume(gomock.Any(), resetToken, time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC)).
		Return(storedAda().ID, true, nil)
	doubles.password.EXPECT().Run(gomock.Any(), adasNewPassword())

	_, err := doubles.resetPassword(t).Run(t.Context(), resetWith(resetToken))

	require.NoError(t, err)
}

func TestResetPasswordRefusesATokenItCannotConsume(t *testing.T) {
	t.Parallel()

	doubles := newResetPasswordDoubles(t)
	doubles.tokens.EXPECT().Consume(gomock.Any(), gomock.Any(), gomock.Any()).Return(domain.UserID(""), false, nil)

	_, err := doubles.resetPassword(t).Run(t.Context(), resetWith("USEDTOKEN"))

	require.ErrorIs(t, err, auth.ErrInvalidPasswordResetToken)
}

func TestResetPasswordFailsWhenThePasswordCannotBeChanged(t *testing.T) {
	t.Parallel()

	doubles := newResetPasswordDoubles(t)
	doubles.tokens.EXPECT().Consume(gomock.Any(), gomock.Any(), gomock.Any()).Return(storedAda().ID, true, nil)
	doubles.password.EXPECT().Run(gomock.Any(), gomock.Any()).Return(auth.ChangePasswordOut{}, errStore)

	_, err := doubles.resetPassword(t).Run(t.Context(), resetWith(resetToken))

	require.ErrorIs(t, err, errStore)
}
