package bootstrap

import (
	"errors"
	"fmt"
	"slices"
)

type Environment string

const (
	EnvironmentLocal      Environment = "local"
	EnvironmentProduction Environment = "production"
)

var errUnknownEnvironment = errors.New("unknown environment")

func environments() []Environment {
	return []Environment{EnvironmentLocal, EnvironmentProduction}
}

func (e *Environment) UnmarshalText(text []byte) error {
	candidate := Environment(text)
	if !slices.Contains(environments(), candidate) {
		return fmt.Errorf("%w %q, want one of %q", errUnknownEnvironment, text, environments())
	}

	*e = candidate

	return nil
}
