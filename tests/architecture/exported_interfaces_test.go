package architecture_test

import (
	"testing"

	"github.com/DannyFestor/go-template-html/tests/architecture/internal/rule"
	"github.com/DannyFestor/go-template-html/tests/architecture/internal/source"
)

func TestNoExportedInterfacesInInternal(t *testing.T) {
	t.Parallel()

	allowlist := map[string]string{}

	requireNoViolations(t, exportedInterfacesInInternal(loadModule(t)), allowlist)
}

func TestNoExportedInterfacesInInternalCatchesFixture(t *testing.T) {
	t.Parallel()

	violations := exportedInterfacesInInternal(loadFixture(t, "exportedinterfaces"))

	requireSubjects(t, violations, "internal/web.Renderer", "internal/web.Number")
}

// Every consumer declares its own unexported interfaces (ADR-0002).
func exportedInterfacesInInternal(files []source.File) []rule.Violation {
	return rule.ExportedInterfaces(source.Under(files, "internal"))
}
