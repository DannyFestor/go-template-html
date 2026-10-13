package commandrule

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/decision"
)

type addsDependency func(args []string) bool

const addCommand = "add"

func addsDependencyCheck(name string) (addsDependency, bool) {
	switch name {
	case "go":
		return goAdds, true
	case "pnpm", "vp":
		return func(args []string) bool { return firstOperand(args) == addCommand }, true
	case "npm":
		return npmAdds, true
	default:
		return nil, false
	}
}

func dependencyAdd(command invocation) decision.Decision {
	adds, known := addsDependencyCheck(filepath.Base(command.Name()))
	if !known || !command.inCheckout(command.Dir) || !adds(command.Args[1:]) {
		return decision.Passed()
	}

	return decision.Asked("`" + strings.Join(command.Args, " ") + "` adds a dependency.")
}

func goAdds(args []string) bool {
	operands := plainSpec().scan(args).operands
	if len(operands) > 0 && operands[0] == "get" {
		return true
	}

	isModEdit := len(operands) >= 2 && operands[0] == "mod" && operands[1] == "edit"

	return isModEdit && slices.ContainsFunc(args, isGoModuleAddOption)
}

func isGoModuleAddOption(arg string) bool {
	name, _, _ := strings.Cut(arg, "=")

	return slices.Contains([]string{"-require", "--require", "-tool", "--tool"}, name)
}

func npmAdds(args []string) bool {
	operands := plainSpec().scan(args).operands

	return len(operands) > 1 && slices.Contains([]string{"install", "i", addCommand}, operands[0])
}

func firstOperand(args []string) string {
	operands := plainSpec().scan(args).operands
	if len(operands) == 0 {
		return ""
	}

	return operands[0]
}
