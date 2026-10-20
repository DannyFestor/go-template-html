package auth

import "context"

//go:generate go tool mockgen -source=device_rememberer.go -destination=mock_device_rememberer_test.go -package=auth -mock_names=deviceRememberer=MockDeviceRememberer

type deviceRememberer interface {
	Run(ctx context.Context, in RememberDeviceIn) (RememberDeviceOut, error)
}
