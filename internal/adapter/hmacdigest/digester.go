package hmacdigest

import (
	"crypto/hmac"
	"crypto/sha256"
)

type Digester struct {
	key []byte
}

func NewDigester(key []byte) *Digester {
	return &Digester{key: key}
}

func (d *Digester) Digest(message []byte) []byte {
	mac := hmac.New(sha256.New, d.key)
	mac.Write(message)

	return mac.Sum(nil)
}
