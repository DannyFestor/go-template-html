package mail

import (
	"context"

	"github.com/DannyFestor/go-template-html/internal/action/mail"
)

//go:generate go tool mockgen -source=mail_sender.go -destination=mock_mail_sender_test.go -package=mail -mock_names=mailSender=MockMailSender
type mailSender interface {
	Run(ctx context.Context, in mail.SendMailIn) (mail.SendMailOut, error)
}
