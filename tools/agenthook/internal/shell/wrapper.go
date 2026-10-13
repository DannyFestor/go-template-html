package shell

import (
	"path/filepath"
	"slices"
	"strings"
)

func unwrap(args []string) []string {
	for len(args) > 0 && isWrapper(args[0]) {
		args = args[1:]
		for len(args) > 0 && isWrapperOption(args[0]) {
			args = args[1:]
		}
	}

	return args
}

func isWrapper(name string) bool {
	return slices.Contains([]string{"env", "command", "exec", "nohup", "builtin"}, filepath.Base(name))
}

func isWrapperOption(arg string) bool {
	return strings.HasPrefix(arg, "-") || strings.Contains(arg, "=")
}
