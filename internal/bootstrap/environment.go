package bootstrap

import (
	"errors"
	"fmt"
)

type Environment string

const (
	EnvironmentLocal      Environment = "local"
	EnvironmentProduction Environment = "production"
)

var errUnknownEnvironment = errors.New("unknown environment")

func (e *Environment) UnmarshalText(text []byte) error {
	candidate := Environment(text)
	if candidate != EnvironmentLocal && candidate != EnvironmentProduction {
		return fmt.Errorf("%w %q, want %q or %q", errUnknownEnvironment, text, EnvironmentLocal, EnvironmentProduction)
	}

	*e = candidate

	return nil
}
