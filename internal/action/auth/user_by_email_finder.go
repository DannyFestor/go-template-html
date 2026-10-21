package auth

import (
	"context"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

//go:generate go tool mockgen -source=user_by_email_finder.go -destination=mock_user_by_email_finder_test.go -package=auth -mock_names=userByEmailFinder=MockUserByEmailFinder

type userByEmailFinder interface {
	FindByEmail(ctx context.Context, email string) (domain.User, bool, error)
}
