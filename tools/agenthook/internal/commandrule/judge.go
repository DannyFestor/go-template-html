package commandrule

import (
	"slices"

	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/decision"
	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/shell"
)

type checkouts interface {
	Root(path string) (string, bool)
	Branch(dir string) string
}

type invocation struct {
	shell.Command

	checkouts checkouts
}

func (i invocation) inCheckout(path string) bool {
	_, found := i.checkouts.Root(path)

	return path != "" && found
}

type rule func(invocation) decision.Decision

type Judge struct {
	checkouts checkouts
	environ   []string
}

func NewJudge(checkouts checkouts, environ []string) Judge {
	return Judge{checkouts: checkouts, environ: environ}
}

func (j Judge) Script(script, dir string) decision.Decision {
	commands, err := shell.Parse(script, shell.Start{Dir: dir, Environ: j.environ, Substitute: j.substitute})
	if err != nil {
		// Invalid shell passes, because it cannot run.
		return decision.Passed()
	}

	var decisions []decision.Decision

	for _, command := range commands {
		for _, rule := range []rule{shellFileWrite, pushToMain, destructiveGit, dependencyAdd} {
			decisions = append(decisions, rule(invocation{Command: command, checkouts: j.checkouts}))
		}
	}

	return decision.Strongest(decisions...)
}

// Agents reach the checkout root through these two, and the checkouts answer both without running the script.
func (j Judge) substitute(command shell.Command) (string, bool) {
	if command.Name() == "pwd" {
		return command.Dir, command.Dir != ""
	}

	git, isGit := parseGit(command)
	if isGit && git.subcommand == "rev-parse" && slices.Equal(git.args, []string{"--show-toplevel"}) {
		return j.checkouts.Root(git.dir)
	}

	return "", false
}
