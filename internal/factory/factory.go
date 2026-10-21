package factory

import (
	"context"
	"sync"

	"github.com/brianvoe/gofakeit/v7"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

const Password = "password"

type userCreator interface {
	Create(ctx context.Context, user domain.NewUser) (domain.User, error)
}

type passwordHasher interface {
	Hash(password string) string
}

type Factory struct {
	users              userCreator
	faker              *gofakeit.Faker
	sharedPasswordHash func() domain.PasswordHash
}

func New(users userCreator, passwords passwordHasher, seed uint64) *Factory {
	return &Factory{
		users: users,
		faker: gofakeit.New(seed),
		sharedPasswordHash: sync.OnceValue(func() domain.PasswordHash {
			return domain.PasswordHash(passwords.Hash(Password))
		}),
	}
}

func (f *Factory) User() *UserBuilder {
	return &UserBuilder{factory: f, name: f.faker.Name(), email: f.faker.Email(), verified: true}
}
