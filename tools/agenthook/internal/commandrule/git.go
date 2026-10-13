package commandrule

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/shell"
)

// A gitInvocation is a git command with its global options applied: dir follows every -C.
type gitInvocation struct {
	dir        string
	subcommand string
	args       []string
}

func parseGit(command shell.Command) (gitInvocation, bool) {
	if filepath.Base(command.Name()) != "git" {
		return gitInvocation{dir: "", subcommand: "", args: nil}, false
	}

	git := gitInvocation{dir: command.Dir, subcommand: "", args: nil}
	args := command.Args[1:]

	for len(args) > 0 && strings.HasPrefix(args[0], "-") {
		args = git.applyGlobalOption(args)
	}

	if len(args) == 0 {
		return git, false
	}

	git.subcommand, git.args = args[0], args[1:]

	return git, true
}

func (g *gitInvocation) applyGlobalOption(args []string) []string {
	valueOptions := []string{"-C", "-c", "--git-dir", "--work-tree", "--namespace", "--config-env"}
	if !slices.Contains(valueOptions, args[0]) || len(args) == 1 {
		return args[1:]
	}

	if args[0] == "-C" {
		g.dir = shell.Resolve(g.dir, args[1])
	}

	return args[2:]
}
