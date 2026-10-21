package auth

import (
	"context"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

//go:generate go tool mockgen -source=remembered_devices_forgetter.go -destination=mock_remembered_devices_forgetter_test.go -package=auth -mock_names=rememberedDevicesForgetter=MockRememberedDevicesForgetter

type rememberedDevicesForgetter interface {
	ForgetDevicesOf(ctx context.Context, userID domain.UserID) error
}
