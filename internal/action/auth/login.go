package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

//nolint:gosec // Hashed once as a decoy, never a credential.
const timingDecoyPassword = "verified in place of an unknown user's password"

var ErrInvalidCredentials = errors.New("the credentials match no user")

type Login struct {
	users     credentialStore
	passwords passwordChecker
	remember  deviceRememberer
	decoyHash domain.PasswordHash
}

type LoginIn struct {
	Email    string
	Password string
	Remember bool
}

type LoginOut struct {
	User          domain.User
	RememberToken domain.RememberToken
}

func NewLogin(users credentialStore, passwords passwordChecker, remember deviceRememberer) *Login {
	return &Login{
		users:     users,
		passwords: passwords,
		remember:  remember,
		decoyHash: domain.PasswordHash(passwords.Hash(timingDecoyPassword)),
	}
}

func (l *Login) Run(ctx context.Context, in LoginIn) (LoginOut, error) {
	loggedIn, err := l.logIn(ctx, in)
	if err != nil {
		return LoginOut{}, fmt.Errorf("log in %s: %w", in.Email, err)
	}

	return loggedIn, nil
}

func (l *Login) logIn(ctx context.Context, in LoginIn) (LoginOut, error) {
	user, err := l.authenticate(ctx, in)
	if err != nil {
		return LoginOut{}, err
	}

	user, err = l.rehash(ctx, user, in.Password)
	if err != nil {
		return LoginOut{}, err
	}

	token, err := l.rememberIfAsked(ctx, user, in.Remember)
	if err != nil {
		return LoginOut{}, err
	}

	return LoginOut{User: user, RememberToken: token}, nil
}

func (l *Login) rememberIfAsked(ctx context.Context, user domain.User, asked bool) (domain.RememberToken, error) {
	var none domain.RememberToken

	if !asked {
		return none, nil
	}

	remembered, err := l.remember.Run(ctx, RememberDeviceIn{UserID: user.ID})
	if err != nil {
		return none, fmt.Errorf("remember the device: %w", err)
	}

	return remembered.Token, nil
}

func (l *Login) authenticate(ctx context.Context, in LoginIn) (domain.User, error) {
	user, found, err := l.users.FindByEmail(ctx, in.Email)
	if err != nil {
		return domain.User{}, fmt.Errorf("find the user: %w", err)
	}

	hash := l.decoyHash
	if found {
		hash = user.PasswordHash
	}

	matches, err := l.passwords.Verify(in.Password, string(hash))
	if err != nil {
		return domain.User{}, fmt.Errorf("verify the password: %w", err)
	}

	if !found || !matches {
		return domain.User{}, ErrInvalidCredentials
	}

	return user, nil
}

func (l *Login) rehash(ctx context.Context, user domain.User, password string) (domain.User, error) {
	if !l.passwords.NeedsRehash(string(user.PasswordHash)) {
		return user, nil
	}

	user.PasswordHash = domain.PasswordHash(l.passwords.Hash(password))

	err := l.users.UpdatePasswordHash(ctx, user.ID, user.PasswordHash)
	if err != nil {
		return domain.User{}, fmt.Errorf("store the rehashed password: %w", err)
	}

	return user, nil
}
