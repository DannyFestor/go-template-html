package signedurl

import (
	"encoding/hex"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Signer struct {
	base  url.URL
	mac   mac
	clock clock
	ttl   time.Duration
}

func NewSigner(base url.URL, key []byte, clock clock, ttl time.Duration) *Signer {
	return &Signer{base: base, mac: mac{key: key}, clock: clock, ttl: ttl}
}

func (s *Signer) Sign(path string, query url.Values) string {
	return s.build(path, cloneQuery(query))
}

func (s *Signer) SignTemporary(path string, query url.Values) string {
	return s.SignTemporaryFor(path, query, s.ttl)
}

func (s *Signer) SignTemporaryFor(path string, query url.Values, ttl time.Duration) string {
	expiring := cloneQuery(query)
	expiring.Set(expiresParam, strconv.FormatInt(s.clock.Now().Add(ttl).Unix(), 10))

	return s.build(path, expiring)
}

func (s *Signer) build(path string, query url.Values) string {
	signed := s.base.JoinPath(path)
	signed.Path = rooted(signed.Path)
	query.Set(signatureParam, hex.EncodeToString(s.mac.sum(signed.Path, query)))
	signed.RawQuery = query.Encode()

	return signed.String()
}

func rooted(path string) string {
	return "/" + strings.TrimPrefix(path, "/")
}

func cloneQuery(query url.Values) url.Values {
	clone := url.Values{}
	for name, values := range query {
		clone[name] = append([]string(nil), values...)
	}

	return clone
}
