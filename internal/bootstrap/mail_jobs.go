package bootstrap

import (
	"context"

	"github.com/DannyFestor/go-template-html/internal/action/mail"
	"github.com/DannyFestor/go-template-html/internal/adapter/clock"
	"github.com/DannyFestor/go-template-html/internal/adapter/smtp"
	"github.com/DannyFestor/go-template-html/internal/domain"
	mailjob "github.com/DannyFestor/go-template-html/internal/job/mail"
	"github.com/DannyFestor/go-template-html/internal/worker"
)

type mailSender interface {
	Send(ctx context.Context, message domain.MailMessage) error
}

func newSMTPSender(config Config) *smtp.Sender {
	return smtp.NewSender(smtpConfig(config.App, config.Mail), clock.System{})
}

func registerMailJobs(jobs *worker.Registry, sender mailSender) {
	worker.Register(jobs, mailjob.NewSendMailHandler(mail.NewSendMail(sender)), worker.DefaultSettings())
}
