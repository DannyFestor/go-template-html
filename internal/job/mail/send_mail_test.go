package mail_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/action/mail"
	"github.com/DannyFestor/go-template-html/internal/domain"
	mailjob "github.com/DannyFestor/go-template-html/internal/job/mail"
)

var errSend = errors.New("send failed")

func sampleMessage() domain.MailMessage {
	return domain.MailMessage{To: "ada@example.test", Subject: "Sample", HTML: "<p>Sample</p>", Text: "Sample"}
}

func TestTheSendMailJobSendsItsMessage(t *testing.T) {
	t.Parallel()

	send := mailjob.NewMockMailSender(gomock.NewController(t))
	send.EXPECT().Run(gomock.Any(), mail.SendMailIn{Message: sampleMessage()})

	err := mailjob.NewSendMailHandler(send).Handle(t.Context(), mailjob.SendMail{Message: sampleMessage()})

	require.NoError(t, err)
}

func TestAFailedSendFailsTheJobSoItIsRetried(t *testing.T) {
	t.Parallel()

	send := mailjob.NewMockMailSender(gomock.NewController(t))
	send.EXPECT().Run(gomock.Any(), mail.SendMailIn{Message: sampleMessage()}).Return(mail.SendMailOut{}, errSend)

	err := mailjob.NewSendMailHandler(send).Handle(t.Context(), mailjob.SendMail{Message: sampleMessage()})

	require.ErrorIs(t, err, errSend)
}
