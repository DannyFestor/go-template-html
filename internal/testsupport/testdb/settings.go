package testdb

import (
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/testsupport/testenv"
)

type settings struct {
	host     string
	port     int
	user     string
	password string
	name     string
	schema   string
}

func settingsFromEnv() (settings, error) {
	values, err := testenv.Required("DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "DB_SCHEMA")
	if err != nil {
		return settings{}, fmt.Errorf("read the Postgres settings: %w", err)
	}

	port, err := values.Port("DB_PORT")
	if err != nil {
		return settings{}, fmt.Errorf("read the Postgres settings: %w", err)
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
