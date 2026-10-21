package auth

import "github.com/DannyFestor/go-template-html/internal/domain"

//go:generate go tool mockgen -source=password_reset_linker.go -destination=mock_password_reset_linker_test.go -package=auth -mock_names=passwordResetLinker=MockPasswordResetLinker

type passwordResetLinker interface {
	PasswordResetURL(token domain.PasswordResetToken) string
}
