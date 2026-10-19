package seed_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/seed"
)

var errDatabaseDown = errors.New("database down")

type recordingRegister struct {
	registered []auth.RegisterIn
	err        error
}

func (r *recordingRegister) Run(_ context.Context, in auth.RegisterIn) (auth.RegisterOut, error) {
	r.registered = append(r.registered, in)

	var user domain.User

	return auth.RegisterOut{User: user}, r.err
}

type recordingUsers struct {
	created []domain.NewUser
}

func (r *recordingUsers) Create(_ context.Context, user domain.NewUser) (domain.User, error) {
	r.created = append(r.created, user)

	var created domain.User

	return created, nil
}

type plainHasher struct{}

func (plainHasher) Hash(password string) string {
	return "hash of " + password
}

func TestSeedingRegistersTheTestUserAndAddsDummyUsers(t *testing.T) {
	t.Parallel()

	register := &recordingRegister{registered: nil, err: nil}
	users := &recordingUsers{created: nil}

	require.NoError(t, seed.New(register, users, plainHasher{}).Run(t.Context()))

	assert.Equal(t, []auth.RegisterIn{{Name: "Test User", Email: "test@example.com", Password: "password"}},
		register.registered)
	assert.Len(t, users.created, 10)
}

func TestSeedingAddsTheSameDummyUsersEveryTime(t *testing.T) {
	t.Parallel()

	first, second := &recordingUsers{created: nil}, &recordingUsers{created: nil}

	require.NoError(t, seed.New(&recordingRegister{registered: nil, err: nil}, first, plainHasher{}).Run(t.Context()))
	require.NoError(t, seed.New(&recordingRegister{registered: nil, err: nil}, second, plainHasher{}).Run(t.Context()))

	assert.Equal(t, first.created, second.created)
}

func TestSeedingStopsWhenTheTestUserCannotBeRegistered(t *testing.T) {
	t.Parallel()

	users := &recordingUsers{created: nil}

	err := seed.New(&recordingRegister{registered: nil, err: errDatabaseDown}, users, plainHasher{}).Run(t.Context())

	require.ErrorIs(t, err, errDatabaseDown)
	assert.Empty(t, users.created)
}
