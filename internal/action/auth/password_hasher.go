package auth

//go:generate go tool mockgen -source=password_hasher.go -destination=mock_password_hasher_test.go -package=auth -mock_names=passwordHasher=MockPasswordHasher

type passwordHasher interface {
	Hash(password string) string
}
