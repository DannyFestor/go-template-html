package check

func AfterEdit(root, rel string) Plan {
	switch {
	case isGo(rel):
		module := goModule(root, rel)
		files := []string{moduleRel(module, rel)}

		return Plan{
			targetCheck(goTarget("fmt", module, files)),
			targetCheck(goTarget("vet", module, files)),
			targetCheck(goTarget("lint-wip", module, files)),
		}
	case isTypeScript(rel):
		return Plan{
			targetCheck(checkoutTarget("fmt-js", []string{rel})),
			targetCheck(checkoutTarget("lint-js", []string{rel})),
		}
	default:
		return nil
	}
}
