package architecture_test

import (
	"testing"

	"github.com/DannyFestor/go-template-html/tests/architecture/internal/rule"
	"github.com/DannyFestor/go-template-html/tests/architecture/internal/source"
)

func TestActionShape(t *testing.T) {
	t.Parallel()

	allowlist := map[string]string{}

	requireNoViolations(t, actionShapeInActions(loadModule(t)), allowlist)
}

func TestActionShapeCatchesFixture(t *testing.T) {
	t.Parallel()

	violations := actionShapeInActions(loadFixture(t, "actionshape"))

	requireSubjects(t, violations,
		"internal/action/auth.Logout",
		"internal/action/auth.Verify",
		"internal/action/auth.Confirm",
		"internal/action/auth.Enable",
		"internal/action/auth.Disable",
		"internal/action/auth.Resend",
		"internal/action/auth.Purge",
	)
}

// One shape for every action lets a single generic decorator wrap them all (ADR-0002).
func actionShapeInActions(files []source.File) []rule.Violation {
	return rule.ActionShape(source.Under(files, "internal/action"))
}
