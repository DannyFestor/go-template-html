package seed_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/seed"
)

var errDatabaseDown = errors.New("database down")

type recordingUsers struct {
	created []domain.NewUser
	err     error
}

func (r *recordingUsers) Create(_ context.Context, user domain.NewUser) (domain.User, error) {
	r.created = append(r.created, user)

	var created domain.User

	return created, r.err
}

type plainHasher struct{}

func (plainHasher) Hash(password string) string {
	return "hash of " + password
}

func TestSeedingAddsTheVerifiedTestUserAndDummyUsers(t *testing.T) {
	t.Parallel()

	users := &recordingUsers{created: nil, err: nil}

	require.NoError(t, seed.New(users, plainHasher{}).Run(t.Context()))

	require.Len(t, users.created, 11)
	assert.Equal(t, domain.NewUser{
		Name:          "Test User",
		Email:         "test@example.com",
		PasswordHash:  "hash of password",
		EmailVerified: true,
	}, users.created[0])
}

func TestSeedingAddsTheSameDummyUsersEveryTime(t *testing.T) {
	t.Parallel()

	first, second := &recordingUsers{created: nil, err: nil}, &recordingUsers{created: nil, err: nil}

	require.NoError(t, seed.New(first, plainHasher{}).Run(t.Context()))
	require.NoError(t, seed.New(second, plainHasher{}).Run(t.Context()))

	assert.Equal(t, first.created, second.created)
}

func TestSeedingStopsWhenTheTestUserCannotBeCreated(t *testing.T) {
	t.Parallel()

	users := &recordingUsers{created: nil, err: errDatabaseDown}

	err := seed.New(users, plainHasher{}).Run(t.Context())

	require.ErrorIs(t, err, errDatabaseDown)
	assert.Len(t, users.created, 1)
}
