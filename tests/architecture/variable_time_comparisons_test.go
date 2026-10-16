package architecture_test

import (
	"testing"

	"github.com/DannyFestor/go-template-html/tests/architecture/internal/rule"
	"github.com/DannyFestor/go-template-html/tests/architecture/internal/source"
)

func TestNoVariableTimeByteComparisonsInInternal(t *testing.T) {
	t.Parallel()

	allowlist := map[string]string{
		"internal/migrationsum.Verify": "compares a checksum of the migrations in the repo, which is no secret",
	}

	requireNoViolations(t, variableTimeComparisonsInInternal(loadModule(t)), allowlist)
}

func TestNoVariableTimeByteComparisonsInInternalCatchesFixture(t *testing.T) {
	t.Parallel()

	violations := variableTimeComparisonsInInternal(loadFixture(t, "variabletimecomparisons"))

	requireSubjects(t, violations,
		"internal/adapter/token.Checker.ByEqual",
		"internal/adapter/token.ByCompare",
		"internal/adapter/token.ByStringConversion",
	)
}

func variableTimeComparisonsInInternal(files []source.File) []rule.Violation {
	return rule.VariableTimeComparisons(source.Under(files, "internal"))
}
