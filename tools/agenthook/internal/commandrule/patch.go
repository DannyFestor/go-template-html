package commandrule

import (
	"path/filepath"
	"slices"
)

func patchDir(command invocation) string {
	if filepath.Base(command.Name()) == "patch" {
		return patchCommandDir(command)
	}

	git, isGit := parseGit(command.Command)
	if !isGit || git.subcommand != "apply" || slices.ContainsFunc(git.args, isReadOnlyApplyOption) {
		return ""
	}

	return git.dir
}

func patchCommandDir(command invocation) string {
	spec := argSpec{
		valueShorts:    "dpioBrDFVYzgx",
		attachedShorts: "",
		valueLongs:     []string{"--directory", "--input", "--output", "--strip"},
	}

	args := spec.scan(command.Args[1:])
	if args.has("--dry-run") {
		return ""
	}

	if args.has("-d", "--directory") {
		return command.Resolve(args.value("-d", "--directory"))
	}

	return command.Dir
}

func isReadOnlyApplyOption(arg string) bool {
	return slices.Contains([]string{"--check", "--stat", "--numstat", "--summary"}, arg)
}
