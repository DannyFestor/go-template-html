package check

func AfterEdit(root, rel string) Plan {
	switch {
	case isGo(rel):
		module := goModule(root, rel)
		files := []string{moduleRel(module, rel)}

		return append(Plan{targetCheck(goTarget("fmt-go", module, files))}, goPackageChecks(module, files)...)
	case isTempl(rel):
		generated := generatedFromTempl(rel)
		module := goModule(root, generated)
		templFiles := []string{rel}

		return append(
			Plan{
				targetCheck(checkoutTarget("fmt-templ", templFiles)),
				targetCheck(checkoutTarget("generate-templ", templFiles)),
			},
			goPackageChecks(module, []string{moduleRel(module, generated)})...,
		)
	case isTypeScript(rel):
		return Plan{
			targetCheck(checkoutTarget("fmt-js", []string{rel})),
			targetCheck(checkoutTarget("lint-js", []string{rel})),
		}
	case isMigration(rel):
		return Plan{
			targetCheck(checkoutTarget("fmt-sql", []string{rel})),
			targetCheck(checkoutTarget("generate-migrations-sum", nil)),
		}
	default:
		return nil
	}
}

func goPackageChecks(module string, files []string) Plan {
	return Plan{
		targetCheck(goTarget("vet", module, files)),
		targetCheck(goTarget("lint-wip", module, files)),
	}
}
