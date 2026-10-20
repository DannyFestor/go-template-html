package auth_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

type requestPasswordResetDoubles struct {
	users  *auth.MockUserByEmailFinder
	issue  *auth.MockPasswordResetLinkIssuer
	mailer *auth.MockResetPasswordMailer
}

func newRequestPasswordResetDoubles(t *testing.T) requestPasswordResetDoubles {
	t.Helper()

	return requestPasswordResetDoubles{
		users:  auth.NewMockUserByEmailFinder(gomock.NewController(t)),
		issue:  auth.NewMockPasswordResetLinkIssuer(gomock.NewController(t)),
		mailer: auth.NewMockResetPasswordMailer(gomock.NewController(t)),
	}
}

func (d requestPasswordResetDoubles) requestPasswordReset(t *testing.T) *auth.RequestPasswordReset {
	t.Helper()

	return auth.NewRequestPasswordReset(d.users, inTransaction(t), d.issue, d.mailer)
}

func (d requestPasswordResetDoubles) adaIsFoundAndIssuedALink() {
	d.users.EXPECT().FindByEmail(gomock.Any(), "ADA@example.com").Return(storedAda(), true, nil)
	d.issue.EXPECT().
		Run(gomock.Any(), auth.IssuePasswordResetLinkIn{UserID: storedAda().ID}).
		Return(auth.IssuePasswordResetLinkOut{URL: resetURL}, nil)
}

func inTransaction(t *testing.T) *auth.MockTransactor {
	t.Helper()

	transactor := auth.NewMockTransactor(gomock.NewController(t))
	transactor.EXPECT().
		WithinTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }).
		AnyTimes()

	return transactor
}

func resetRequestFor(email string) auth.RequestPasswordResetIn {
	return auth.RequestPasswordResetIn{Email: email}
}

func TestRequestPasswordResetMailsTheUserAPasswordResetLink(t *testing.T) {
	t.Parallel()

	doubles := newRequestPasswordResetDoubles(t)
	doubles.adaIsFoundAndIssuedALink()
	doubles.mailer.EXPECT().Send(gomock.Any(), domain.ResetPasswordMail{To: "ada@example.com", URL: resetURL})

	_, err := doubles.requestPasswordReset(t).Run(t.Context(), resetRequestFor("ADA@example.com"))

	require.NoError(t, err)
}

func TestRequestPasswordResetForAnUnknownEmailSucceedsWithoutMailing(t *testing.T) {
	t.Parallel()

	var unknown domain.User

	doubles := newRequestPasswordResetDoubles(t)
	doubles.users.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(unknown, false, nil)

	_, err := doubles.requestPasswordReset(t).Run(t.Context(), resetRequestFor("grace@example.com"))

	require.NoError(t, err)
}

func TestRequestPasswordResetFailsWhenTheMailCannotBeQueued(t *testing.T) {
	t.Parallel()

	doubles := newRequestPasswordResetDoubles(t)
	doubles.adaIsFoundAndIssuedALink()
	doubles.mailer.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errStore)

	_, err := doubles.requestPasswordReset(t).Run(t.Context(), resetRequestFor("ADA@example.com"))

	require.ErrorIs(t, err, errStore)
}
