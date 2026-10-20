package factory

import (
	"context"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type UserBuilder struct {
	factory *Factory
	name    string
	email   string
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.email = email

	return b
}

func (b *UserBuilder) Create(ctx context.Context) (domain.User, error) {
	user, err := b.factory.users.Create(ctx, domain.NewUser{
		Name:         b.name,
		Email:        b.email,
		PasswordHash: b.factory.sharedPasswordHash(),
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("create the factory user %s: %w", b.email, err)
	}

	return user, nil
}
