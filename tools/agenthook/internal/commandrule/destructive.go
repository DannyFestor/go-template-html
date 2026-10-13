package commandrule

import (
	"slices"

	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/decision"
)

type discardsWork func(scannedArgs) bool

func discardsWorkCheck(subcommand string) (discardsWork, bool) {
	switch subcommand {
	case "reset":
		return func(args scannedArgs) bool { return args.has("--hard") }, true
	case "clean":
		return func(args scannedArgs) bool { return args.has("-f", "--force") }, true
	case "restore":
		return func(args scannedArgs) bool { return !args.has("-S", "--staged") }, true
	case "checkout":
		return checkoutDiscards, true
	case "stash":
		return stashDiscards, true
	case "branch":
		return branchForceDeletes, true
	case "worktree":
		return worktreeForceRemoves, true
	default:
		return nil, false
	}
}

func destructiveGit(command invocation) decision.Decision {
	git, isGit := parseGit(command.Command)
	if !isGit || !command.inCheckout(git.dir) {
		return decision.Passed()
	}

	discards, known := discardsWorkCheck(git.subcommand)
	if !known || !discards(scanGitArgs(git.args)) {
		return decision.Passed()
	}

	return decision.Asked("`git " + git.subcommand + "` discards work git cannot recover.")
}

// The operands after -- are kept apart, because checkout -- <path> discards while checkout <branch> does not.
func scanGitArgs(args []string) scannedArgs {
	spec := argSpec{valueShorts: "es", attachedShorts: "", valueLongs: []string{"--source"}}

	separator := slices.Index(args, "--")
	if separator < 0 {
		return spec.scan(args)
	}

	scanned := spec.scan(args[:separator])
	scanned.options["--"] = ""
	scanned.operands = append(scanned.operands, args[separator+1:]...)

	return scanned
}

func checkoutDiscards(args scannedArgs) bool {
	return (args.has("--") && len(args.operands) > 0) || slices.Contains(args.operands, ".")
}

func stashDiscards(args scannedArgs) bool {
	return len(args.operands) > 0 && (args.operands[0] == "drop" || args.operands[0] == "clear")
}

func branchForceDeletes(args scannedArgs) bool {
	return args.has("-D") || (args.has("-d", "--delete") && args.has("-f", "--force"))
}

func worktreeForceRemoves(args scannedArgs) bool {
	return len(args.operands) > 0 && args.operands[0] == "remove" && args.has("-f", "--force")
}
