package bootstrap

import (
	"github.com/DannyFestor/go-template-html/internal/adapter/clock"
	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/adapter/signedurl"
	"github.com/DannyFestor/go-template-html/internal/adapter/verificationlink"
	"github.com/DannyFestor/go-template-html/internal/domain"
	mailjob "github.com/DannyFestor/go-template-html/internal/job/mail"
	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/view/mail"
)

func newVerificationLinks(config Config, signedURLKey []byte) *verificationlink.Links {
	signer := signedurl.NewSigner(config.App.URL, signedURLKey, clock.System{}, config.SignedURL.TTL)

	return verificationlink.NewLinks(signer, route.VerifyEmail, config.SignedURL.EmailVerificationTTL)
}

func newVerificationMailer(jobs *postgres.JobQueue) *mailjob.Mailer[domain.VerifyEmailMail] {
	return mailjob.NewMailer(composeVerifyEmail, jobs)
}

func composeVerifyEmail(verification domain.VerifyEmailMail) mailjob.Content {
	return mailjob.Content{
		To:      verification.To,
		Subject: mail.VerifyEmailSubject,
		HTML:    mail.Layout(mail.VerifyEmailSubject, mail.VerifyEmail(verification.URL)),
		Text:    mail.VerifyEmailText(verification.URL),
	}
}
