package check

type runner interface {
	Run(root string, target Target) Run
}

type check interface {
	run(maker runner, root string) (failure string)
}

type Plan []check

// FirstFailure stops at the first failing check, because later ones mostly fail for the same cause.
func (p Plan) FirstFailure(maker runner, root string) string {
	for _, check := range p {
		if failure := check.run(maker, root); failure != "" {
			return failure
		}
	}

	return ""
}

type targetCheck Target

func (t targetCheck) run(maker runner, root string) string {
	run := maker.Run(root, Target(t))
	if !run.Failed {
		return ""
	}

	return failure(run)
}

func failure(run Run) string {
	return "`" + run.Target.String() + "` failed:\n" + run.Output()
}
