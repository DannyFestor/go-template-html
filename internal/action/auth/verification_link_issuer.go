package auth

import "github.com/DannyFestor/go-template-html/internal/domain"

//go:generate go tool mockgen -source=verification_link_issuer.go -destination=mock_verification_link_issuer_test.go -package=auth -mock_names=verificationLinkIssuer=MockVerificationLinkIssuer

type verificationLinkIssuer interface {
	For(user domain.User) string
}
