package auth

import "context"

//go:generate go tool mockgen -source=password_reset_link_issuer.go -destination=mock_password_reset_link_issuer_test.go -package=auth -mock_names=passwordResetLinkIssuer=MockPasswordResetLinkIssuer

type passwordResetLinkIssuer interface {
	Run(ctx context.Context, in IssuePasswordResetLinkIn) (IssuePasswordResetLinkOut, error)
}
