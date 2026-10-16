package signedurl

import (
	"crypto/hmac"
	"crypto/sha256"
	"net/url"
)

type mac struct {
	key []byte
}

func (m mac) sum(path string, query url.Values) []byte {
	hash := hmac.New(sha256.New, m.key)
	hash.Write([]byte(canonical(path, query)))

	return hash.Sum(nil)
}

func (m mac) matches(path string, query url.Values, signature []byte) bool {
	return hmac.Equal(m.sum(path, query), signature)
}

func canonical(path string, query url.Values) string {
	return path + "?" + query.Encode()
}
