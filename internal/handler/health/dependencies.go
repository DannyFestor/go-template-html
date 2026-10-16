package health

import (
	"context"
	"errors"
)

type Dependencies struct {
	pingers []pinger
}

func NewDependencies(pingers ...pinger) *Dependencies {
	return &Dependencies{pingers: pingers}
}

func (d *Dependencies) Ping(ctx context.Context) error {
	errs := make([]error, 0, len(d.pingers))
	for _, dependency := range d.pingers {
		errs = append(errs, dependency.Ping(ctx))
	}

	return errors.Join(errs...)
}
