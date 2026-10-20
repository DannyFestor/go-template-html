package auth

import (
	"context"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

//go:generate go tool mockgen -source=credential_store.go -destination=mock_credential_store_test.go -package=auth -mock_names=credentialStore=MockCredentialStore

type credentialStore interface {
	FindByEmail(ctx context.Context, email string) (domain.User, bool, error)
	UpdatePasswordHash(ctx context.Context, id domain.UserID, hash domain.PasswordHash) error
}
