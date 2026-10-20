package auth

import (
	"context"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

//go:generate go tool mockgen -source=user_finder.go -destination=mock_user_finder_test.go -package=auth -mock_names=userFinder=MockUserFinder

type userFinder interface {
	Find(ctx context.Context, id domain.UserID) (domain.User, bool, error)
}
