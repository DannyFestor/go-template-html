package testdb

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

var errUnset = errors.New("unset: run the tests through make, which passes the Compose Postgres settings")

type settings struct {
	host     string
	port     int
	user     string
	password string
	name     string
	schema   string
}

func settingsFromEnv() (settings, error) {
	values := map[string]string{}

	var missing []error

	for _, key := range []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SCHEMA"} {
		values[key] = os.Getenv(key)
		if values[key] == "" {
			missing = append(missing, fmt.Errorf("%s %w", key, errUnset))
		}
	}

	if len(missing) > 0 {
		return settings{}, errors.Join(missing...)
	}

	port, err := strconv.Atoi(values["DB_PORT"])
	if err != nil {
		return settings{}, fmt.Errorf("DB_PORT: %w", err)
	}

	return settings{
		host:     values["DB_HOST"],
		port:     port,
		user:     values["DB_USER"],
		password: values["DB_PASSWORD"],
		name:     values["DB_NAME"],
		schema:   values["DB_SCHEMA"],
	}, nil
}

func (s settings) database(name string) Database {
	return Database{Host: s.host, Port: s.port, User: s.user, Password: s.password, Name: name, Schema: s.schema}
}
