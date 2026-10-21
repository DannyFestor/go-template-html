package bootstrap

import (
	"time"

	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/domain"
	mailjob "github.com/DannyFestor/go-template-html/internal/job/mail"
	"github.com/DannyFestor/go-template-html/internal/view/mail"
)

func newResetPasswordMailer(jobs *postgres.JobQueue, ttl time.Duration) *mailjob.Mailer[domain.ResetPasswordMail] {
	return mailjob.NewMailer(composeResetPasswordMail(ttl), jobs)
}

func composeResetPasswordMail(ttl time.Duration) func(domain.ResetPasswordMail) mailjob.Content {
	return func(sent domain.ResetPasswordMail) mailjob.Content {
		model := mail.ResetPassword{URL: sent.URL, ExpiresInMinutes: int(ttl.Minutes())}

		return mailjob.Content{
			To:      sent.To,
			Subject: mail.ResetPasswordSubject,
			HTML:    mail.Layout(mail.ResetPasswordSubject, mail.ResetPasswordBody(model)),
			Text:    mail.ResetPasswordText(model),
		}
	}
}
