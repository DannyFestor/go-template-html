package argon2id

import (
	"errors"
	"fmt"
)

var errZeroCost = errors.New("argon2id needs at least one iteration and one lane")

type Params struct {
	MemoryKiB   uint32
	Iterations  uint32
	Parallelism uint8
}

func (p Params) checkPositiveCost() error {
	if p.Iterations == 0 || p.Parallelism == 0 {
		return fmt.Errorf("%w: t=%d, p=%d", errZeroCost, p.Iterations, p.Parallelism)
	}

	return nil
}
