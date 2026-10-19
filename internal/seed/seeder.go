package seed

import (
	"context"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/factory"
)

const (
	testUserName  = "Test User"
	testUserEmail = "test@example.com"
	dummyUsers    = 10
	fakerSeed     = 1
)

type registrar interface {
	Run(ctx context.Context, in auth.RegisterIn) (auth.RegisterOut, error)
}

type userCreator interface {
	Create(ctx context.Context, user domain.NewUser) (domain.User, error)
}

type passwordHasher interface {
	Hash(password string) string
}

type Seeder struct {
	register registrar
	factory  *factory.Factory
}

func New(register registrar, users userCreator, passwords passwordHasher) *Seeder {
	return &Seeder{register: register, factory: factory.New(users, passwords, fakerSeed)}
}

func (s *Seeder) Run(ctx context.Context) error {
	_, err := s.register.Run(ctx, auth.RegisterIn{Name: testUserName, Email: testUserEmail, Password: factory.Password})
	if err != nil {
		return fmt.Errorf("seed the test user: %w", err)
	}

	return s.seedDummyUsers(ctx)
}

func (s *Seeder) seedDummyUsers(ctx context.Context) error {
	for range dummyUsers {
		_, err := s.factory.User().Create(ctx)
		if err != nil {
			return fmt.Errorf("seed the dummy users: %w", err)
		}
	}

	return nil
}
