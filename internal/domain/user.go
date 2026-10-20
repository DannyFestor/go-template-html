package domain

import "time"

type UserID string

type PasswordHash string

type User struct {
	ID              UserID
	Name            string
	Email           string
	PasswordHash    PasswordHash
	EmailVerifiedAt time.Time
	CreatedAt       time.Time
}

func (u User) Verified() bool {
	return !u.EmailVerifiedAt.IsZero()
}
