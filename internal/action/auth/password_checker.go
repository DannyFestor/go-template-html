package auth

//go:generate go tool mockgen -source=password_checker.go -destination=mock_password_checker_test.go -package=auth -mock_names=passwordChecker=MockPasswordChecker

type passwordChecker interface {
	Hash(password string) string
	Verify(password, encoded string) (bool, error)
	NeedsRehash(encoded string) bool
}
