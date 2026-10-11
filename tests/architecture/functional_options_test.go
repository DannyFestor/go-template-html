package architecture_test

import (
	"testing"

	"github.com/DannyFestor/go-template-html/tests/architecture/internal/rule"
)

func TestNoFunctionalOptions(t *testing.T) {
	t.Parallel()

	allowlist := map[string]string{}

	requireNoViolations(t, rule.FunctionalOptions(loadModule(t)), allowlist)
}

func TestNoFunctionalOptionsCatchesFixture(t *testing.T) {
	t.Parallel()

	violations := rule.FunctionalOptions(loadFixture(t, "functionaloptions"))

	requireSubjects(t, violations,
		"internal/adapter/smtp.Option",
		"internal/adapter/smtp.New",
		"internal/adapter/smtp.Client.With",
		"cmd/web.setting",
	)
}
