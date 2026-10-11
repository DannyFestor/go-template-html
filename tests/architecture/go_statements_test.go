package architecture_test

import (
	"testing"

	"github.com/DannyFestor/go-template-html/tests/architecture/internal/rule"
	"github.com/DannyFestor/go-template-html/tests/architecture/internal/source"
)

func TestNoGoStatementsInCoreLayers(t *testing.T) {
	t.Parallel()

	allowlist := map[string]string{}

	requireNoViolations(t, goStatementsInCoreLayers(loadModule(t)), allowlist)
}

func TestNoGoStatementsInCoreLayersCatchesFixture(t *testing.T) {
	t.Parallel()

	violations := goStatementsInCoreLayers(loadFixture(t, "gostatements"))

	requireSubjects(t, violations, "internal/action/spawn.Start", "internal/action/spawn.Spawner.Run")
}

// Asynchrony belongs to adapters, decorators and bootstrap (ADR-0002).
func goStatementsInCoreLayers(files []source.File) []rule.Violation {
	coreLayers := []string{
		"internal/domain",
		"internal/action",
		"internal/request",
		"internal/validate",
		"internal/handler",
		"internal/middleware",
	}

	return rule.GoStatements(source.Under(files, coreLayers...))
}
