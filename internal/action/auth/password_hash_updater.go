package auth

import (
	"context"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

//go:generate go tool mockgen -source=password_hash_updater.go -destination=mock_password_hash_updater_test.go -package=auth -mock_names=passwordHashUpdater=MockPasswordHashUpdater

type passwordHashUpdater interface {
	UpdatePasswordHash(ctx context.Context, id domain.UserID, hash domain.PasswordHash) error
}
