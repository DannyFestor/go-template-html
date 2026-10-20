package auth

//go:generate go tool mockgen -source=verification_link_matcher.go -destination=mock_verification_link_matcher_test.go -package=auth -mock_names=verificationLinkMatcher=MockVerificationLinkMatcher

type verificationLinkMatcher interface {
	Matches(email, hash string) bool
}
