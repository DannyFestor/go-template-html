package auth

import (
	"context"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

//go:generate go tool mockgen -source=user_creator.go -destination=mock_user_creator_test.go -package=auth -mock_names=userCreator=MockUserCreator

type userCreator interface {
	Create(ctx context.Context, user domain.NewUser) (domain.User, error)
}
