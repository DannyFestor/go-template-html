package auth

import "time"

//go:generate go tool mockgen -source=clock.go -destination=mock_clock_test.go -package=auth -mock_names=clock=MockClock

type clock interface {
	Now() time.Time
}
