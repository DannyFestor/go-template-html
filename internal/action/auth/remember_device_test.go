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

const rememberLifetime = 30 * 24 * time.Hour

func fixedClock(t *testing.T) *auth.MockClock {
	t.Helper()

	clock := auth.NewMockClock(gomock.NewController(t))
	clock.EXPECT().Now().Return(time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC)).AnyTimes()

	return clock
}

func issuedToken() domain.RememberToken {
	return domain.RememberToken{Selector: "SELECTOR2", Validator: "VALIDATOR2"}
}

func TestRememberDeviceIssuesATokenThatExpiresAfterTheRememberLifetime(t *testing.T) {
	t.Parallel()

	generator := auth.NewMockRememberTokenGenerator(gomock.NewController(t))
	generator.EXPECT().New().Return(issuedToken())

	tokens := auth.NewMockRememberTokenIssuer(gomock.NewController(t))
	tokens.EXPECT().
		Issue(gomock.Any(), storedAda().ID, issuedToken(), time.Date(2026, 11, 20, 9, 0, 0, 0, time.UTC)).
		Return(nil)

	remembered, err := auth.NewRememberDevice(generator, tokens, fixedClock(t), rememberLifetime).
		Run(t.Context(), auth.RememberDeviceIn{UserID: storedAda().ID})

	require.NoError(t, err)
	assert.Equal(t, auth.RememberDeviceOut{Token: issuedToken()}, remembered)
}
