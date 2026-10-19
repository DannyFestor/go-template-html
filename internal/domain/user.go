package domain

import "time"

type UserID string

type PasswordHash string

type User struct {
	ID           UserID
	Name         string
	Email        string
	PasswordHash PasswordHash
	CreatedAt    time.Time
}
