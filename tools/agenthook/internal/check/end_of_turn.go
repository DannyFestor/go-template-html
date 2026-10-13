package check

import (
	"maps"
	"slices"
)

// EndOfTurn plans the checks for the files one session edited in a checkout, rel to its root. Each Go module is
// built, linted and tested on the edited packages alone, so another session's work in the same checkout cannot
// fail this one.
func EndOfTurn(root string, editedPaths []string) Plan {
	var plan Plan

	filesByModule := goFilesByModule(root, editedPaths)
	for _, module := range slices.Sorted(maps.Keys(filesByModule)) {
		plan = append(plan, goModulePlan(module, filesByModule[module])...)
	}

	if slices.ContainsFunc(editedPaths, isTypeScript) {
		plan = append(
			plan,
			targetCheck(checkoutTarget("typecheck-js", nil)),
			targetCheck(checkoutTarget("test-js", nil)),
		)
	}

	return plan
}

// go-arch-lint only knows the root module, and checks all of it, so its warnings are cut down to the files.
func goModulePlan(module string, files []string) Plan {
	plan := Plan{targetCheck(goTarget("build-packages", module, files))}
	if module == "." {
		plan = append(plan, archCheck{files: files})
	}

	return append(plan, targetCheck(goTarget("lint", module, files)), targetCheck(goTarget("test-go", module, files)))
}

func goFilesByModule(root string, editedPaths []string) map[string][]string {
	filesByModule := map[string][]string{}

	for _, rel := range editedPaths {
		if isGo(rel) {
			module := goModule(root, rel)
			filesByModule[module] = append(filesByModule[module], moduleRel(module, rel))
		}
	}

	return filesByModule
}
