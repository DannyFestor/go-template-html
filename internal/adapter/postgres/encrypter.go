package postgres

type encrypter interface {
	Encrypt(plaintext []byte) []byte
	Decrypt(ciphertext []byte) ([]byte, error)
}
