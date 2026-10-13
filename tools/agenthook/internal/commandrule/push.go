package commandrule

import (
	"strings"

	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/decision"
)

const protectedBranch = "main"

func pushToMain(command invocation) decision.Decision {
	git, isGit := parseGit(command.Command)
	if !isGit || git.subcommand != "push" || !pushesProtectedBranch(git, command.checkouts) {
		return decision.Passed()
	}

	return decision.Denied("Pushes to " + protectedBranch + " are refused: push a branch and open a pull request.")
}

// Without a refspec, git pushes the branch checked out where it runs.
func pushesProtectedBranch(git gitInvocation, checkouts checkouts) bool {
	spec := argSpec{
		valueShorts:    "o",
		attachedShorts: "",
		valueLongs:     []string{"--push-option", "--repo", "--receive-pack", "--exec"},
	}

	args := spec.scan(git.args)
	if args.has("--all", "--mirror") {
		return true
	}

	currentBranch := checkouts.Branch(git.dir)

	refspecs := refspecsOf(args.operands)
	if len(refspecs) == 0 {
		return currentBranch == protectedBranch
	}

	for _, refspec := range refspecs {
		if refspecDestination(refspec, currentBranch) == protectedBranch {
			return true
		}
	}

	return false
}

// The first operand of git push is the remote.
func refspecsOf(operands []string) []string {
	if len(operands) == 0 {
		return nil
	}

	return operands[1:]
}

func refspecDestination(refspec, currentBranch string) string {
	refspec = strings.TrimPrefix(refspec, "+")
	if _, destination, mapped := strings.Cut(refspec, ":"); mapped {
		refspec = destination
	}

	if refspec == "HEAD" {
		return currentBranch
	}

	return strings.TrimPrefix(refspec, "refs/heads/")
}
