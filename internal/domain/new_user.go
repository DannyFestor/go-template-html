package domain

import "errors"

var ErrEmailTaken = errors.New("the email address is already taken")

type NewUser struct {
	Name          string
	Email         string
	PasswordHash  PasswordHash
	EmailVerified bool
}
