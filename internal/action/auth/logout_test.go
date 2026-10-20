package auth_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
)

func TestLogoutForgetsTheRememberTokenOfTheDevice(t *testing.T) {
	t.Parallel()

	tokens := auth.NewMockRememberTokenForgetter(gomock.NewController(t))
	tokens.EXPECT().Forget(gomock.Any(), presentedToken()).Return(nil)

	_, err := auth.NewLogout(tokens).Run(t.Context(), auth.LogoutIn{RememberToken: presentedToken()})

	require.NoError(t, err)
}
