package token

import (
	"bytes"
	"crypto/hmac"
	"crypto/subtle"
)

type Checker struct{ want []byte }

func (c Checker) ByEqual(got []byte) bool {
	return bytes.Equal(c.want, got)
}

func ByCompare(want, got []byte) int {
	return bytes.Compare(want, got)
}

func ByStringConversion(want, got []byte) bool {
	return string(want) != string(got)
}

func ByHMAC(want, got []byte) bool {
	return hmac.Equal(want, got)
}

func BySubtle(want, got []byte) bool {
	return subtle.ConstantTimeCompare(want, got) == 1
}

func ByLength(want, got []byte) bool {
	return len(want) == len(got)
}
