package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

const storedHash domain.PasswordHash = "$argon2id$v=19$m=8,t=1,p=1$c2FsdA$a2V5"

func registration() auth.RegisterIn {
	return auth.RegisterIn{Name: "Ada Lovelace", Email: "ada@example.com", Password: "correct horse"}
}

func newRegister(t *testing.T, users *auth.MockUserCreator) *auth.Register {
	t.Helper()

	passwords := auth.NewMockPasswordHasher(gomock.NewController(t))
	passwords.EXPECT().Hash("correct horse").Return(string(storedHash)).AnyTimes()

	return auth.NewRegister(users, passwords)
}

func TestRegisterCreatesTheUserWithTheHashOfThePassword(t *testing.T) {
	t.Parallel()

	created := domain.User{
		ID:           "019a1f6e-3c4b-7d2e-8f10-1234567890ab",
		Name:         "Ada Lovelace",
		Email:        "ada@example.com",
		PasswordHash: storedHash,
		CreatedAt:    time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC),
	}
	users := auth.NewMockUserCreator(gomock.NewController(t))
	users.EXPECT().
		Create(gomock.Any(), domain.NewUser{
			Name:         "Ada Lovelace",
			Email:        "ada@example.com",
			PasswordHash: storedHash,
		}).
		Return(created, nil)

	registered, err := newRegister(t, users).Run(t.Context(), registration())

	require.NoError(t, err)
	assert.Equal(t, auth.RegisterOut{User: created}, registered)
}

func TestRegisterReportsATakenEmail(t *testing.T) {
	t.Parallel()

	var none domain.User

	users := auth.NewMockUserCreator(gomock.NewController(t))
	users.EXPECT().Create(gomock.Any(), gomock.Any()).Return(none, domain.ErrEmailTaken)

	_, err := newRegister(t, users).Run(t.Context(), registration())

	require.ErrorIs(t, err, domain.ErrEmailTaken)
}
