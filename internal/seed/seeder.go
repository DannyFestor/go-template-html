package seed

import (
	"context"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/factory"
)

const (
	testUserName  = "Test User"
	testUserEmail = "test@example.com"
	dummyUsers    = 10
	fakerSeed     = 1
)

type userCreator interface {
	Create(ctx context.Context, user domain.NewUser) (domain.User, error)
}

type passwordHasher interface {
	Hash(password string) string
}

type Seeder struct {
	factory *factory.Factory
}

func New(users userCreator, passwords passwordHasher) *Seeder {
	return &Seeder{factory: factory.New(users, passwords, fakerSeed)}
}

func (s *Seeder) Run(ctx context.Context) error {
	_, err := s.factory.User().WithName(testUserName).WithEmail(testUserEmail).Create(ctx)
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
