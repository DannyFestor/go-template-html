package auth

import "github.com/DannyFestor/go-template-html/internal/domain"

//go:generate go tool mockgen -source=password_reset_token_generator.go -destination=mock_password_reset_token_generator_test.go -package=auth -mock_names=passwordResetTokenGenerator=MockPasswordResetTokenGenerator

type passwordResetTokenGenerator interface {
	New() domain.PasswordResetToken
}
