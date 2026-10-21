package auth

import (
	"context"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

//go:generate go tool mockgen -source=verification_mailer.go -destination=mock_verification_mailer_test.go -package=auth -mock_names=verificationMailer=MockVerificationMailer

type verificationMailer interface {
	Send(ctx context.Context, mail domain.VerifyEmailMail) error
}
