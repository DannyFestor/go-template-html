package auth

import (
	"context"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

//go:generate go tool mockgen -source=reset_password_mailer.go -destination=mock_reset_password_mailer_test.go -package=auth -mock_names=resetPasswordMailer=MockResetPasswordMailer

type resetPasswordMailer interface {
	Send(ctx context.Context, mail domain.ResetPasswordMail) error
}
