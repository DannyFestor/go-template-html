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

const (
	dummyHash   = "$argon2id$v=19$m=8,t=1,p=1$ZHVtbXk$a2V5"
	rehashed    = "$argon2id$v=19$m=16,t=1,p=1$bmV3$a2V5"
	rightSecret = "correct horse"
	wrongSecret = "battery staple"
)

type loginDoubles struct {
	users     *auth.MockCredentialStore
	passwords *auth.MockPasswordChecker
	remember  *auth.MockDeviceRememberer
}

func newLoginDoubles(t *testing.T) loginDoubles {
	t.Helper()

	passwords := auth.NewMockPasswordChecker(gomock.NewController(t))
	passwords.EXPECT().Hash(gomock.Any()).Return(dummyHash)

	return loginDoubles{
		users:     auth.NewMockCredentialStore(gomock.NewController(t)),
		passwords: passwords,
		remember:  auth.NewMockDeviceRememberer(gomock.NewController(t)),
	}
}

func (d loginDoubles) login() *auth.Login {
	return auth.NewLogin(d.users, d.passwords, d.remember)
}

func (d loginDoubles) adaLogsIn() {
	d.users.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(storedAda(), true, nil)
	d.passwords.EXPECT().Verify(rightSecret, string(storedHash)).Return(true, nil)
	d.passwords.EXPECT().NeedsRehash(string(storedHash)).Return(false)
}

func storedAda() domain.User {
	return domain.User{
		ID:           "019a1f6e-3c4b-7d2e-8f10-1234567890ab",
		Name:         "Ada Lovelace",
		Email:        "ada@example.com",
		PasswordHash: storedHash,
		CreatedAt:    time.Date(2026, 10, 20, 9, 0, 0, 0, time.UTC),
	}
}

func credentials(password string) auth.LoginIn {
	return auth.LoginIn{Email: "ADA@example.com", Password: password, Remember: false}
}

func TestLoginReturnsTheUserWhosePasswordMatches(t *testing.T) {
	t.Parallel()

	doubles := newLoginDoubles(t)
	doubles.adaLogsIn()

	loggedIn, err := doubles.login().Run(t.Context(), credentials(rightSecret))

	require.NoError(t, err)
	assert.Equal(t, storedAda(), loggedIn.User)
	assert.Zero(t, loggedIn.RememberToken)
}

func TestLoginRemembersTheDeviceWhenAsked(t *testing.T) {
	t.Parallel()

	doubles := newLoginDoubles(t)
	doubles.adaLogsIn()
	doubles.remember.EXPECT().
		Run(gomock.Any(), auth.RememberDeviceIn{UserID: storedAda().ID}).
		Return(auth.RememberDeviceOut{Token: issuedToken()}, nil)

	remembered := credentials(rightSecret)
	remembered.Remember = true

	loggedIn, err := doubles.login().Run(t.Context(), remembered)

	require.NoError(t, err)
	assert.Equal(t, auth.LoginOut{User: storedAda(), RememberToken: issuedToken()}, loggedIn)
}

func TestLoginRefusesAWrongPassword(t *testing.T) {
	t.Parallel()

	doubles := newLoginDoubles(t)
	doubles.users.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(storedAda(), true, nil)
	doubles.passwords.EXPECT().Verify(wrongSecret, string(storedHash)).Return(false, nil)

	_, err := doubles.login().Run(t.Context(), credentials(wrongSecret))

	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestLoginVerifiesAnUnknownEmailAgainstTheDummyHashAndRefusesIt(t *testing.T) {
	t.Parallel()

	var unknown domain.User

	doubles := newLoginDoubles(t)
	doubles.users.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(unknown, false, nil)
	doubles.passwords.EXPECT().Verify(rightSecret, dummyHash).Return(true, nil)

	_, err := doubles.login().Run(t.Context(), credentials(rightSecret))

	require.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestLoginRehashesAPasswordHashedWithOutdatedParameters(t *testing.T) {
	t.Parallel()

	doubles := newLoginDoubles(t)
	doubles.users.EXPECT().FindByEmail(gomock.Any(), gomock.Any()).Return(storedAda(), true, nil)
	doubles.passwords.EXPECT().Verify(rightSecret, string(storedHash)).Return(true, nil)
	doubles.passwords.EXPECT().NeedsRehash(string(storedHash)).Return(true)
	doubles.passwords.EXPECT().Hash(rightSecret).Return(rehashed)
	doubles.users.EXPECT().UpdatePasswordHash(gomock.Any(), storedAda().ID, domain.PasswordHash(rehashed)).Return(nil)

	loggedIn, err := doubles.login().Run(t.Context(), credentials(rightSecret))

	require.NoError(t, err)
	assert.Equal(t, domain.PasswordHash(rehashed), loggedIn.User.PasswordHash)
}
