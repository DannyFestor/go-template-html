package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/action/auth"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

const storedHash domain.PasswordHash = "$argon2id$v=19$m=8,t=1,p=1$c2FsdA$a2V5"

type registerDoubles struct {
	users        *auth.MockUserCreator
	verification *auth.MockVerificationSender
}

func newRegisterDoubles(t *testing.T) registerDoubles {
	t.Helper()

	return registerDoubles{
		users:        auth.NewMockUserCreator(gomock.NewController(t)),
		verification: auth.NewMockVerificationSender(gomock.NewController(t)),
	}
}

func (d registerDoubles) register(t *testing.T) *auth.Register {
	t.Helper()

	passwords := auth.NewMockPasswordHasher(gomock.NewController(t))
	passwords.EXPECT().Hash("correct horse").Return(string(storedHash)).AnyTimes()

	transactor := auth.NewMockTransactor(gomock.NewController(t))
	transactor.EXPECT().
		WithinTransaction(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })

	return auth.NewRegister(transactor, d.users, passwords, d.verification)
}

func registration() auth.RegisterIn {
	return auth.RegisterIn{Name: "Ada Lovelace", Email: "ada@example.com", Password: "correct horse"}
}

func createdAda() domain.User {
	return domain.User{
		ID:              "019a1f6e-3c4b-7d2e-8f10-1234567890ab",
		Name:            "Ada Lovelace",
		Email:           "ada@example.com",
		PasswordHash:    storedHash,
		EmailVerifiedAt: time.Time{},
		CreatedAt:       time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC),
	}
}

func TestRegisterCreatesAnUnverifiedUserWithTheHashOfThePasswordAndMailsTheVerificationLink(t *testing.T) {
	t.Parallel()

	doubles := newRegisterDoubles(t)
	doubles.users.EXPECT().
		Create(gomock.Any(), domain.NewUser{
			Name:          "Ada Lovelace",
			Email:         "ada@example.com",
			PasswordHash:  storedHash,
			EmailVerified: false,
		}).
		Return(createdAda(), nil)
	doubles.verification.EXPECT().Run(gomock.Any(), auth.SendVerificationMailIn{User: createdAda()})

	registered, err := doubles.register(t).Run(t.Context(), registration())

	require.NoError(t, err)
	assert.Equal(t, auth.RegisterOut{User: createdAda()}, registered)
}

func TestRegisterReportsATakenEmailWithoutMailing(t *testing.T) {
	t.Parallel()

	var none domain.User

	doubles := newRegisterDoubles(t)
	doubles.users.EXPECT().Create(gomock.Any(), gomock.Any()).Return(none, domain.ErrEmailTaken)

	_, err := doubles.register(t).Run(t.Context(), registration())

	require.ErrorIs(t, err, domain.ErrEmailTaken)
}

func TestRegisterFailsWhenTheVerificationMailCannotBeSent(t *testing.T) {
	t.Parallel()

	doubles := newRegisterDoubles(t)
	doubles.users.EXPECT().Create(gomock.Any(), gomock.Any()).Return(createdAda(), nil)
	doubles.verification.EXPECT().
		Run(gomock.Any(), gomock.Any()).
		Return(auth.SendVerificationMailOut{}, errQueueDown)

	_, err := doubles.register(t).Run(t.Context(), registration())

	require.ErrorIs(t, err, errQueueDown)
}
