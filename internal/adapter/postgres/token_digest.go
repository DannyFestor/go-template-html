package postgres

import "crypto/sha256"

func tokenDigest(token string) []byte {
	digest := sha256.Sum256([]byte(token))

	return digest[:]
}
