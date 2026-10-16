package signedurl

import (
	"encoding/hex"
	"errors"
	"net/url"
	"strconv"
	"time"
)

var (
	ErrInvalidSignature = errors.New("the URL's signature is missing or does not match")
	ErrExpired          = errors.New("the signed URL has expired")
)

type Verifier struct {
	mac   mac
	clock clock
}

func NewVerifier(key []byte, clock clock) *Verifier {
	return &Verifier{mac: mac{key: key}, clock: clock}
}

func (v *Verifier) Verify(target *url.URL) error {
	query := target.Query()

	signature, err := hex.DecodeString(query.Get(signatureParam))
	query.Del(signatureParam)

	if err != nil || !v.mac.matches(target.Path, query, signature) {
		return ErrInvalidSignature
	}

	return v.checkExpiry(query)
}

func (v *Verifier) checkExpiry(query url.Values) error {
	if !query.Has(expiresParam) {
		return nil
	}

	expires, err := strconv.ParseInt(query.Get(expiresParam), 10, 64)
	if err != nil {
		return ErrInvalidSignature
	}

	if !v.clock.Now().Before(time.Unix(expires, 0)) {
		return ErrExpired
	}

	return nil
}
