package testenv

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

var errUnset = errors.New("unset: run the tests through make, which passes the Compose service settings")

type Values map[string]string

func Required(keys ...string) (Values, error) {
	values := Values{}

	var missing []error

	for _, key := range keys {
		values[key] = os.Getenv(key)
		if values[key] == "" {
			missing = append(missing, fmt.Errorf("%s %w", key, errUnset))
		}
	}

	if len(missing) > 0 {
		return nil, errors.Join(missing...)
	}

	return values, nil
}

func (v Values) Port(key string) (int, error) {
	port, err := strconv.Atoi(v[key])
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}

	return port, nil
}
