package factory_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/factory"
)

const seed = 42

var errStoreDown = errors.New("store down")

type recordingUsers struct {
	created []domain.NewUser
	err     error
}

func (r *recordingUsers) Create(_ context.Context, user domain.NewUser) (domain.User, error) {
	r.created = append(r.created, user)

	return domain.User{
		ID:              "019a1f6e-3c4b-7d2e-8f10-1234567890ab",
		Name:            user.Name,
		Email:           user.Email,
		PasswordHash:    user.PasswordHash,
		EmailVerifiedAt: time.Time{},
		CreatedAt:       time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC),
	}, r.err
}

type countingHasher struct {
	hashed []string
}

func (c *countingHasher) Hash(password string) string {
	c.hashed = append(c.hashed, password)

	return "hash of " + password
}

func TestAUserGetsAFakeNameAndEmailAndTheHashOfTheFactoryPassword(t *testing.T) {
	t.Parallel()

	users := &recordingUsers{created: nil, err: nil}

	user, err := factory.New(users, &countingHasher{hashed: nil}, seed).User().Create(t.Context())

	require.NoError(t, err)
	assert.NotEmpty(t, user.Name)
	assert.Contains(t, user.Email, "@")
	assert.Equal(t, domain.PasswordHash("hash of "+factory.Password), user.PasswordHash)
}

func TestTheSameSeedMakesTheSameUsers(t *testing.T) {
	t.Parallel()

	first, second := &recordingUsers{created: nil, err: nil}, &recordingUsers{created: nil, err: nil}

	createUsers(t, factory.New(first, &countingHasher{hashed: nil}, seed), 3)
	createUsers(t, factory.New(second, &countingHasher{hashed: nil}, seed), 3)

	assert.Equal(t, first.created, second.created)
}

func TestTheFactoryPasswordIsHashedOnce(t *testing.T) {
	t.Parallel()

	passwords := &countingHasher{hashed: nil}

	createUsers(t, factory.New(&recordingUsers{created: nil, err: nil}, passwords, seed), 3)

	assert.Equal(t, []string{factory.Password}, passwords.hashed)
}

func TestWithEmailOverridesTheFakeEmail(t *testing.T) {
	t.Parallel()

	users := &recordingUsers{created: nil, err: nil}

	user, err := factory.New(users, &countingHasher{hashed: nil}, seed).User().
		WithEmail("ada@example.com").
		Create(t.Context())

	require.NoError(t, err)
	assert.Equal(t, "ada@example.com", user.Email)
}

func TestAUserHasAVerifiedEmailUnlessAskedOtherwise(t *testing.T) {
	t.Parallel()

	users := &recordingUsers{created: nil, err: nil}
	fakes := factory.New(users, &countingHasher{hashed: nil}, seed)

	createUsers(t, fakes, 1)
	_, err := fakes.User().Unverified().Create(t.Context())
	require.NoError(t, err)

	require.Len(t, users.created, 2)
	assert.True(t, users.created[0].EmailVerified, "default")
	assert.False(t, users.created[1].EmailVerified, "unverified")
}

func TestWithNameOverridesTheFakeName(t *testing.T) {
	t.Parallel()

	users := &recordingUsers{created: nil, err: nil}

	user, err := factory.New(users, &countingHasher{hashed: nil}, seed).User().
		WithName("Ada Lovelace").
		Create(t.Context())

	require.NoError(t, err)
	assert.Equal(t, "Ada Lovelace", user.Name)
}

func TestCreateReportsAFailingStore(t *testing.T) {
	t.Parallel()

	users := &recordingUsers{created: nil, err: errStoreDown}

	_, err := factory.New(users, &countingHasher{hashed: nil}, seed).User().Create(t.Context())

	require.ErrorIs(t, err, errStoreDown)
}

func createUsers(t *testing.T, users *factory.Factory, count int) {
	t.Helper()

	for range count {
		_, err := users.User().Create(t.Context())
		require.NoError(t, err)
	}
}
