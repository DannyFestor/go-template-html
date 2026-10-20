package auth

import "context"

//go:generate go tool mockgen -source=verification_sender.go -destination=mock_verification_sender_test.go -package=auth -mock_names=verificationSender=MockVerificationSender

type verificationSender interface {
	Run(ctx context.Context, in SendVerificationMailIn) (SendVerificationMailOut, error)
}
