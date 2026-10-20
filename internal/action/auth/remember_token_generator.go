package auth

import "github.com/DannyFestor/go-template-html/internal/domain"

//go:generate go tool mockgen -source=remember_token_generator.go -destination=mock_remember_token_generator_test.go -package=auth -mock_names=rememberTokenGenerator=MockRememberTokenGenerator

type rememberTokenGenerator interface {
	New() domain.RememberToken
}
