package auth

import (
	"context"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type ChangePassword struct {
	passwords passwordHasher
	users     passwordHashUpdater
	devices   rememberedDevicesForgetter
}

type ChangePasswordIn struct {
	UserID   domain.UserID
	Password string
}

type ChangePasswordOut struct{}

func NewChangePassword(
	passwords passwordHasher,
	users passwordHashUpdater,
	devices rememberedDevicesForgetter,
) *ChangePassword {
	return &ChangePassword{passwords: passwords, users: users, devices: devices}
}

func (c *ChangePassword) Run(ctx context.Context, in ChangePasswordIn) (ChangePasswordOut, error) {
	err := c.change(ctx, in)
	if err != nil {
		return ChangePasswordOut{}, fmt.Errorf("change the password of user %s: %w", in.UserID, err)
	}

	return ChangePasswordOut{}, nil
}

func (c *ChangePassword) change(ctx context.Context, in ChangePasswordIn) error {
	err := c.users.UpdatePasswordHash(ctx, in.UserID, domain.PasswordHash(c.passwords.Hash(in.Password)))
	if err != nil {
		return fmt.Errorf("store the new password hash: %w", err)
	}

	err = c.devices.ForgetDevicesOf(ctx, in.UserID)
	if err != nil {
		return fmt.Errorf("forget the remembered devices: %w", err)
	}

	return nil
}
