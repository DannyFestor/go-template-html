package auth

import "context"

//go:generate go tool mockgen -source=password_changer.go -destination=mock_password_changer_test.go -package=auth -mock_names=passwordChanger=MockPasswordChanger

type passwordChanger interface {
	Run(ctx context.Context, in ChangePasswordIn) (ChangePasswordOut, error)
}
