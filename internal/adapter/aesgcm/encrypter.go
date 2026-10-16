package aesgcm

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
)

const keySize = 32

var (
	ErrUndecryptable = errors.New("the ciphertext is tampered with, truncated or sealed under another key")
	errKeySize       = fmt.Errorf("an AES-256 key must be %d bytes", keySize)
)

type Encrypter struct {
	aead cipher.AEAD
}

func NewEncrypter(key []byte) (*Encrypter, error) {
	if len(key) != keySize {
		return nil, errKeySize
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create the AES cipher: %w", err)
	}

	aead, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return nil, fmt.Errorf("create the GCM mode: %w", err)
	}

	return &Encrypter{aead: aead}, nil
}

func (e *Encrypter) Encrypt(plaintext []byte) []byte {
	//nolint:gosec // NewGCMWithRandomNonce draws the nonce itself and prepends it, so the nonce argument must be nil.
	return e.aead.Seal(nil, nil, plaintext, nil)
}

func (e *Encrypter) Decrypt(ciphertext []byte) ([]byte, error) {
	plaintext, err := e.aead.Open(nil, nil, ciphertext, nil)
	if err != nil {
		return nil, ErrUndecryptable
	}

	return plaintext, nil
}
