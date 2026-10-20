package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
)

func TestPruneExpiredRememberTokensDeletesTheTokensExpiredByNow(t *testing.T) {
	t.Parallel()

	tokens := auth.NewMockExpiredRememberTokenDeleter(gomock.NewController(t))
	tokens.EXPECT().DeleteExpired(gomock.Any(), time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC)).Return(nil)

	_, err := auth.NewPruneExpiredRememberTokens(tokens, fixedClock(t)).
		Run(t.Context(), auth.PruneExpiredRememberTokensIn{})

	require.NoError(t, err)
}
