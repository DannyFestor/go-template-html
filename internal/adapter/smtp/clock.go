package smtp

import "time"

type clock interface {
	Now() time.Time
}
