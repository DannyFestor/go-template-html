package editrule

import "github.com/DannyFestor/go-template-html/tools/agenthook/internal/decision"

func guardrailFile(edit Edit) decision.Decision {
	return askBeforeEditing(edit, []string{
		".golangci.yml",
		".go-arch-lint.yml",
		"lefthook.yml",
		"tools/agenthook/**",
		".claude/settings.json",
		".github/workflows/**",
		".github/rulesets/**",
	}, "is a guardrail file: it configures the checks the agent works under.")
}

func dependencyManifest(edit Edit) decision.Decision {
	return askBeforeEditing(edit, []string{"go.mod", "tools/*.mod", "package.json"},
		"lists dependencies: the edit may add one.")
}

func askBeforeEditing(edit Edit, patterns []string, reason string) decision.Decision {
	if !matchesAny(patterns, edit.Path) {
		return decision.Passed()
	}

	return decision.Asked(edit.Path + " " + reason)
}
