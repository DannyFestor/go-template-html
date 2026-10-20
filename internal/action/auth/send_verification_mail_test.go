package auth_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

const adasVerificationLink = "https://example.com/verify-email/019a1f6e/hash?expires=1&signature=abc"

var errQueueDown = errors.New("queue down")

func newSendVerificationMail(t *testing.T, mailer *auth.MockVerificationMailer) *auth.SendVerificationMail {
	t.Helper()

	links := auth.NewMockVerificationLinkIssuer(gomock.NewController(t))
	links.EXPECT().For(storedAda()).Return(adasVerificationLink).AnyTimes()

	return auth.NewSendVerificationMail(links, mailer)
}

func TestSendVerificationMailMailsTheUserTheirVerificationLink(t *testing.T) {
	t.Parallel()

	mailer := auth.NewMockVerificationMailer(gomock.NewController(t))
	mailer.EXPECT().Send(gomock.Any(), domain.VerifyEmailMail{To: "ada@example.com", URL: adasVerificationLink})

	_, err := newSendVerificationMail(t, mailer).
		Run(t.Context(), auth.SendVerificationMailIn{User: storedAda()})

	require.NoError(t, err)
}

func TestSendVerificationMailFailsWhenTheMailCannotBeSent(t *testing.T) {
	t.Parallel()

	mailer := auth.NewMockVerificationMailer(gomock.NewController(t))
	mailer.EXPECT().Send(gomock.Any(), gomock.Any()).Return(errQueueDown)

	_, err := newSendVerificationMail(t, mailer).
		Run(t.Context(), auth.SendVerificationMailIn{User: storedAda()})

	require.ErrorIs(t, err, errQueueDown)
}
