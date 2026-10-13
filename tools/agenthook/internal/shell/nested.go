package shell

import (
	"path/filepath"
	"slices"
)

func nestedScript(command Command) (string, bool) {
	if !slices.Contains([]string{"sh", "bash", "zsh", "dash"}, filepath.Base(command.Name())) {
		return "", false
	}

	flag := slices.Index(command.Args, "-c")
	if flag >= 0 && flag+1 < len(command.Args) {
		return command.Args[flag+1], true
	}

	return command.Input, command.Input != ""
}
