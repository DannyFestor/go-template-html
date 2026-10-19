package mail_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/action/mail"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

var errDelivery = errors.New("the mail server refused the recipient")

func sampleMessage() domain.MailMessage {
	return domain.MailMessage{To: "ada@example.test", Subject: "Sample", HTML: "<p>Sample</p>", Text: "Sample"}
}

func TestSendMailHandsTheMessageToTheSender(t *testing.T) {
	t.Parallel()

	sender := mail.NewMockSender(gomock.NewController(t))
	sender.EXPECT().Send(gomock.Any(), sampleMessage())

	_, err := mail.NewSendMail(sender).Run(t.Context(), mail.SendMailIn{Message: sampleMessage()})

	require.NoError(t, err)
}

func TestSendMailFailsWhenTheSenderFails(t *testing.T) {
	t.Parallel()

	sender := mail.NewMockSender(gomock.NewController(t))
	sender.EXPECT().Send(gomock.Any(), sampleMessage()).Return(errDelivery)

	_, err := mail.NewSendMail(sender).Run(t.Context(), mail.SendMailIn{Message: sampleMessage()})

	require.ErrorIs(t, err, errDelivery)
}
