package argon2id

import (
	"crypto/rand"
	"crypto/subtle"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	rfc9106SaltBytes = 16
	rfc9106KeyBytes  = 32
)

type Hasher struct {
	params Params
}

func NewHasher(params Params) (*Hasher, error) {
	err := params.checkPositiveCost()
	if err != nil {
		return nil, fmt.Errorf("configure the password hasher: %w", err)
	}

	return &Hasher{params: params}, nil
}

func (h *Hasher) Hash(password string) string {
	salt := make([]byte, rfc9106SaltBytes)
	_, _ = rand.Read(salt)

	return phcHash{params: h.params, salt: salt, key: deriveKey(password, salt, h.params)}.String()
}

func (h *Hasher) Verify(password, encoded string) (bool, error) {
	stored, err := parsePHCHash(encoded)
	if err != nil {
		return false, fmt.Errorf("verify the password: %w", err)
	}

	return subtle.ConstantTimeCompare(deriveKey(password, stored.salt, stored.params), stored.key) == 1, nil
}

func deriveKey(password string, salt []byte, params Params) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.MemoryKiB,
		params.Parallelism,
		rfc9106KeyBytes,
	)
}
