package middleware

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

type sessionToken string

func newSessionToken() sessionToken {
	return sessionToken(rand.Text())
}

// storageID hashes the token, so a key read out of Valkey can't be replayed as a cookie.
func (t sessionToken) storageID() string {
	sum := sha256.Sum256([]byte(t))

	return hex.EncodeToString(sum[:])
}
