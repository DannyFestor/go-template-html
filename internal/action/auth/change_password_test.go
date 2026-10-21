package auth_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

const newHash domain.PasswordHash = "$argon2id$v=19$m=8,t=1,p=1$bmV3$a2V5"

var errStore = errors.New("the store is down")

type changePasswordDoubles struct {
	users   *auth.MockPasswordHashUpdater
	devices *auth.MockRememberedDevicesForgetter
}

func newChangePasswordDoubles(t *testing.T) changePasswordDoubles {
	t.Helper()

	return changePasswordDoubles{
		users:   auth.NewMockPasswordHashUpdater(gomock.NewController(t)),
		devices: auth.NewMockRememberedDevicesForgetter(gomock.NewController(t)),
	}
}

func (d changePasswordDoubles) changePassword(t *testing.T) *auth.ChangePassword {
	t.Helper()

	passwords := auth.NewMockPasswordHasher(gomock.NewController(t))
	passwords.EXPECT().Hash("new password").Return(string(newHash)).AnyTimes()

	return auth.NewChangePassword(passwords, d.users, d.devices)
}

func adasNewPassword() auth.ChangePasswordIn {
	return auth.ChangePasswordIn{UserID: storedAda().ID, Password: "new password"}
}

func TestChangePasswordStoresTheHashOfTheNewPasswordAndForgetsTheRememberedDevices(t *testing.T) {
	t.Parallel()

	doubles := newChangePasswordDoubles(t)
	gomock.InOrder(
		doubles.users.EXPECT().UpdatePasswordHash(gomock.Any(), storedAda().ID, newHash),
		doubles.devices.EXPECT().ForgetDevicesOf(gomock.Any(), storedAda().ID),
	)

	_, err := doubles.changePassword(t).Run(t.Context(), adasNewPassword())

	require.NoError(t, err)
}

func TestChangePasswordFailsWhenTheHashCannotBeStored(t *testing.T) {
	t.Parallel()

	doubles := newChangePasswordDoubles(t)
	doubles.users.EXPECT().UpdatePasswordHash(gomock.Any(), gomock.Any(), gomock.Any()).Return(errStore)

	_, err := doubles.changePassword(t).Run(t.Context(), adasNewPassword())

	require.ErrorIs(t, err, errStore)
}

func TestChangePasswordFailsWhenTheRememberedDevicesCannotBeForgotten(t *testing.T) {
	t.Parallel()

	doubles := newChangePasswordDoubles(t)
	doubles.users.EXPECT().UpdatePasswordHash(gomock.Any(), gomock.Any(), gomock.Any())
	doubles.devices.EXPECT().ForgetDevicesOf(gomock.Any(), gomock.Any()).Return(errStore)

	_, err := doubles.changePassword(t).Run(t.Context(), adasNewPassword())

	require.ErrorIs(t, err, errStore)
}
