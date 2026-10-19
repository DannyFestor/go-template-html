package totp

import "time"

//go:generate go tool mockgen -source=clock.go -destination=mock_clock_test.go -package=totp -mock_names=clock=MockClock

type clock interface {
	Now() time.Time
}
