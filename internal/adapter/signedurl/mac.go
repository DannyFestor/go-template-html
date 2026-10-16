package signedurl

import (
	"crypto/hmac"
	"net/url"

	"github.com/DannyFestor/go-template-html/internal/adapter/hmacdigest"
)

type mac struct {
	digester *hmacdigest.Digester
}

func newMAC(key []byte) mac {
	return mac{digester: hmacdigest.NewDigester(key)}
}

func (m mac) sum(path string, query url.Values) []byte {
	return m.digester.Digest([]byte(canonical(path, query)))
}

func (m mac) matches(path string, query url.Values, signature []byte) bool {
	return hmac.Equal(m.sum(path, query), signature)
}

func canonical(path string, query url.Values) string {
	return path + "?" + query.Encode()
}
