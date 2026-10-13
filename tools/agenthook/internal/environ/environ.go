package environ

import (
	"slices"
	"strings"
)

func Without(environ []string, names ...string) []string {
	return keepNames(environ, func(name string) bool { return !slices.Contains(names, name) })
}

func WithoutPrefix(environ []string, prefix string) []string {
	return keepNames(environ, func(name string) bool { return !strings.HasPrefix(name, prefix) })
}

func keepNames(environ []string, isKept func(name string) bool) []string {
	var kept []string

	for _, variable := range environ {
		name, _, _ := strings.Cut(variable, "=")
		if isKept(name) {
			kept = append(kept, variable)
		}
	}

	return kept
}
