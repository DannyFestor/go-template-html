package middleware

import "time"

type sessionData struct {
	UserID                string    `json:"userId,omitzero"`
	CredentialFingerprint string    `json:"credentialFingerprint,omitzero"`
	Flash                 string    `json:"flash,omitzero"`
	IntendedURL           string    `json:"intendedUrl,omitzero"`
	PasswordConfirmedAt   time.Time `json:"passwordConfirmedAt,omitzero"`
	ExpiresAt             time.Time `json:"expiresAt"`
}

func (d sessionData) isEmpty() bool {
	var empty sessionData

	d.ExpiresAt = empty.ExpiresAt

	return d == empty
}
