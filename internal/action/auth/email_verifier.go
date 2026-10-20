package auth

import (
	"context"
	"time"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

//go:generate go tool mockgen -source=email_verifier.go -destination=mock_email_verifier_test.go -package=auth -mock_names=emailVerifier=MockEmailVerifier

type emailVerifier interface {
	MarkEmailVerified(ctx context.Context, id domain.UserID, at time.Time) error
}
