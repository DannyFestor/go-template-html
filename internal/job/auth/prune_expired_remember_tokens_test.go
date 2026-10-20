package auth_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
	authjob "github.com/DannyFestor/go-template-html/internal/job/auth"
)

var errPrune = errors.New("prune failed")

func TestAFailedPruneFailsTheJobSoItIsRetried(t *testing.T) {
	t.Parallel()

	prune := authjob.NewMockExpiredRememberTokenPruner(gomock.NewController(t))
	prune.EXPECT().
		Run(gomock.Any(), auth.PruneExpiredRememberTokensIn{}).
		Return(auth.PruneExpiredRememberTokensOut{}, errPrune)

	err := authjob.NewPruneExpiredRememberTokensHandler(prune).
		Handle(t.Context(), authjob.PruneExpiredRememberTokens{})

	require.ErrorIs(t, err, errPrune)
}
