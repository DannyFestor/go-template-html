package appkey

import (
	"crypto/hkdf"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const (
	Size       = 32
	SubkeySize = 32
)

var ErrMalformedKey = fmt.Errorf("the app key must be %d bytes in standard base64", Size)

type Key struct {
	material []byte
}

func Parse(encoded string) (Key, error) {
	material, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil || len(material) != Size {
		return Key{material: nil}, ErrMalformedKey
	}

	return Key{material: material}, nil
}

func (k Key) Subkey(purpose Purpose) ([]byte, error) {
	subkey, err := hkdf.Key(sha256.New, k.material, nil, string(purpose), SubkeySize)
	if err != nil {
		return nil, fmt.Errorf("derive the %s subkey: %w", purpose, err)
	}

	return subkey, nil
}
