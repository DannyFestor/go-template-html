package auth

//go:generate go tool mockgen -source=password_verifier.go -destination=mock_password_verifier_test.go -package=auth -mock_names=passwordVerifier=MockPasswordVerifier

type passwordVerifier interface {
	Verify(password, encoded string) (bool, error)
}
