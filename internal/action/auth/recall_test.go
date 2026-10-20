package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

type recallDoubles struct {
	tokens   *auth.MockRememberTokenConsumer
	users    *auth.MockUserFinder
	remember *auth.MockDeviceRememberer
}

func newRecallDoubles(t *testing.T) recallDoubles {
	t.Helper()

	return recallDoubles{
		tokens:   auth.NewMockRememberTokenConsumer(gomock.NewController(t)),
		users:    auth.NewMockUserFinder(gomock.NewController(t)),
		remember: auth.NewMockDeviceRememberer(gomock.NewController(t)),
	}
}

func (d recallDoubles) recall(t *testing.T) *auth.Recall {
	t.Helper()

	transactor := auth.NewMockTransactor(gomock.NewController(t))
	transactor.EXPECT().
		WithinTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })

	return auth.NewRecall(transactor, d.tokens, d.users, d.remember, fixedClock(t))
}

func presentedToken() domain.RememberToken {
	return domain.RememberToken{Selector: "SELECTOR1", Validator: "VALIDATOR1"}
}

func adasUse(replay bool) domain.RememberTokenUse {
	return domain.RememberTokenUse{UserID: storedAda().ID, Replay: replay}
}

func TestRecallConsumesTheTokenAndRemembersTheDeviceWithAFreshOne(t *testing.T) {
	t.Parallel()

	doubles := newRecallDoubles(t)
	doubles.tokens.EXPECT().
		Consume(gomock.Any(), presentedToken(), time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC)).
		Return(adasUse(false), true, nil)
	doubles.users.EXPECT().Find(gomock.Any(), storedAda().ID).Return(storedAda(), true, nil)
	doubles.remember.EXPECT().
		Run(gomock.Any(), auth.RememberDeviceIn{UserID: storedAda().ID}).
		Return(auth.RememberDeviceOut{Token: issuedToken()}, nil)

	recalled, err := doubles.recall(t).Run(t.Context(), auth.RecallIn{Token: presentedToken()})

	require.NoError(t, err)
	assert.Equal(t, auth.RecallOut{User: storedAda(), Token: issuedToken()}, recalled)
}

func TestRecallOfAReplayedTokenRecallsTheUserWithoutANewToken(t *testing.T) {
	t.Parallel()

	doubles := newRecallDoubles(t)
	doubles.tokens.EXPECT().Consume(gomock.Any(), presentedToken(), gomock.Any()).Return(adasUse(true), true, nil)
	doubles.users.EXPECT().Find(gomock.Any(), storedAda().ID).Return(storedAda(), true, nil)

	recalled, err := doubles.recall(t).Run(t.Context(), auth.RecallIn{Token: presentedToken()})

	require.NoError(t, err)
	assert.Equal(t, storedAda(), recalled.User)
	assert.Zero(t, recalled.Token)
}

func TestRecallRefusesATokenItCannotConsume(t *testing.T) {
	t.Parallel()

	var unused domain.RememberTokenUse

	doubles := newRecallDoubles(t)
	doubles.tokens.EXPECT().Consume(gomock.Any(), gomock.Any(), gomock.Any()).Return(unused, false, nil)

	_, err := doubles.recall(t).Run(t.Context(), auth.RecallIn{Token: presentedToken()})

	require.ErrorIs(t, err, auth.ErrNotRemembered)
}

func TestRecallRefusesATokenWhoseUserIsGone(t *testing.T) {
	t.Parallel()

	var gone domain.User

	doubles := newRecallDoubles(t)
	doubles.tokens.EXPECT().Consume(gomock.Any(), gomock.Any(), gomock.Any()).Return(adasUse(false), true, nil)
	doubles.users.EXPECT().Find(gomock.Any(), storedAda().ID).Return(gone, false, nil)

	_, err := doubles.recall(t).Run(t.Context(), auth.RecallIn{Token: presentedToken()})

	require.ErrorIs(t, err, auth.ErrNotRemembered)
}
