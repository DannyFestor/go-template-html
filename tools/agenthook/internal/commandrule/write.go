package commandrule

import (
	"path/filepath"
	"slices"

	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/decision"
)

const shellWriteReason = "Shell writes into the repo are refused, so every change goes through the edit tools and " +
	"their post-edit checks. Use the Edit or Write tool, or write scratch files under $TMPDIR."

type writer func(invocation) []string

const sourceAndDestination = 2

func writerFor(name string) (writer, bool) {
	switch name {
	case "tee":
		return teeTargets, true
	case "cp", "mv", "ln":
		return copyTarget, true
	case "sed":
		return sedEditor().targets, true
	case "perl":
		return perlEditor().targets, true
	default:
		return nil, false
	}
}

func shellFileWrite(command invocation) decision.Decision {
	if slices.ContainsFunc(writtenPaths(command), command.inCheckout) || writesInPlace(command) {
		return decision.Denied(shellWriteReason)
	}

	return decision.Passed()
}

func writtenPaths(command invocation) []string {
	paths := slices.Clone(command.Writes)

	if writer, writes := writerFor(filepath.Base(command.Name())); writes {
		for _, target := range writer(command) {
			paths = append(paths, command.Resolve(target))
		}
	}

	return paths
}

// writesInPlace covers commands that write files they do not name: patches and interpreter scripts.
func writesInPlace(command invocation) bool {
	dir := patchDir(command)
	if dir == "" && interpreterWritesFiles(command) {
		dir = command.Dir
	}

	return command.inCheckout(dir)
}

func teeTargets(command invocation) []string {
	return plainSpec().scan(command.Args[1:]).operands
}

func copyTarget(command invocation) []string {
	spec := argSpec{valueShorts: "tS", attachedShorts: "", valueLongs: []string{"--target-directory", "--suffix"}}

	args := spec.scan(command.Args[1:])
	if args.has("-t", "--target-directory") {
		return []string{args.value("-t", "--target-directory")}
	}

	if len(args.operands) < sourceAndDestination {
		return nil
	}

	return args.operands[len(args.operands)-1:]
}
