package bootstrap

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

func LoadConfig(dotenvPath string, environ []string) (Config, error) {
	values, err := readDotenv(dotenvPath)
	if err != nil {
		return Config{}, err
	}

	maps.Copy(values, env.ToMap(environ))

	config, err := env.ParseAsWithOptions[Config](parseOptions(values))
	if err != nil {
		return Config{}, fmt.Errorf("load config: %w", err)
	}

	return config, nil
}

func parseOptions(environment map[string]string) env.Options {
	return env.Options{
		Environment:                  environment,
		TagName:                      "",
		PrefixTagName:                "",
		DefaultValueTagName:          "",
		RequiredIfNoDef:              false,
		OnSet:                        nil,
		Prefix:                       "",
		UseFieldNameByDefault:        false,
		SetDefaultsForZeroValuesOnly: false,
		FuncMap:                      nil,
	}
}

func readDotenv(path string) (map[string]string, error) {
	values, err := godotenv.Read(path)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]string{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	return values, nil
}
