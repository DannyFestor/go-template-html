package rule

import (
	"go/ast"
	"maps"
	"slices"
	"strings"

	"github.com/DannyFestor/go-template-html/tests/architecture/internal/source"
)

// actionCheck returns why an action breaks its shape, or "" when it holds.
type actionCheck func(pkg packageIndex, action string) string

func ActionShape(files []source.File) []Violation {
	checks := []actionCheck{isStructAction, hasPointerConstructor, hasOnlyRun, hasRunSignature, hasCompanionStructs}

	var violations []Violation

	for _, pkg := range indexPackages(files) {
		for _, action := range actionNames(pkg) {
			declared := pkg.types[action]
			for _, check := range checks {
				if message := check(pkg, action); message != "" {
					violations = append(violations, newViolation(declared.file, declared.spec, action, message))
				}
			}
		}
	}

	return violations
}

func actionNames(pkg packageIndex) []string {
	var names []string

	for _, name := range slices.Sorted(maps.Keys(pkg.types)) {
		if ast.IsExported(name) && !pkg.isCompanion(name) {
			names = append(names, name)
		}
	}

	return names
}

func isStructAction(pkg packageIndex, action string) string {
	if pkg.isStruct(action) {
		return ""
	}

	return "is not a struct"
}

func hasPointerConstructor(pkg packageIndex, action string) string {
	constructor, ok := pkg.funcs["New"+action]
	if !ok {
		return "has no New" + action + " constructor"
	}

	results := fieldTypes(constructor.Type.Results)
	if len(results) != 1 || !isPointerTo(results[0], action) {
		return "New" + action + " must return only *" + action
	}

	return ""
}

func hasOnlyRun(pkg packageIndex, action string) string {
	var extra []string

	for _, method := range pkg.exportedMethods(action) {
		if method.Name.Name != "Run" {
			extra = append(extra, method.Name.Name)
		}
	}

	if len(extra) == 0 {
		return ""
	}

	return "has exported methods besides Run: " + strings.Join(extra, ", ")
}

func hasRunSignature(pkg packageIndex, action string) string {
	for _, method := range pkg.exportedMethods(action) {
		if method.Name.Name == "Run" && isRunSignature(method.Type, action) {
			return ""
		}
	}

	return "needs Run(ctx context.Context, in " + action + "In) (" + action + "Out, error)"
}

func isRunSignature(funcType *ast.FuncType, action string) bool {
	params := fieldTypes(funcType.Params)
	results := fieldTypes(funcType.Results)

	return len(params) == 2 && isSelector(params[0], "context", "Context") && isIdent(params[1], action+"In") &&
		len(results) == 2 && isIdent(results[0], action+"Out") && isIdent(results[1], "error")
}

func hasCompanionStructs(pkg packageIndex, action string) string {
	var missing []string

	for _, companion := range []string{action + "In", action + "Out"} {
		if !pkg.isStruct(companion) {
			missing = append(missing, companion)
		}
	}

	if len(missing) == 0 {
		return ""
	}

	return "needs struct types " + strings.Join(missing, " and ")
}
