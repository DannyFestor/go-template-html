package hooks_test

import "testing"

const (
	coverageSummaryScript  = "coverage-summary.sh"
	coverageProfile        = "coverage.out"
	coverageProfileContent = `mode: atomic
example.com/shop/internal/order/order.go:3.1,5.2 3 4
example.com/shop/internal/order/order.go:7.1,9.2 1 0
example.com/shop/internal/user/user.go:3.1,5.2 2 0
`
)

func TestCoverageSummaryReportsTheTotalAndEachPackage(t *testing.T) {
	t.Parallel()

	module := moduleWithCoverage(t)

	run := runScript(t, module, "", coverageSummaryScript, coverageProfile, "40")

	requirePassed(t, run)
	requireOutput(t, run, "**50.0%**")
	requireOutput(t, run, "| `internal/order` | 75.0% |")
	requireOutput(t, run, "| `internal/user` | 0.0% |")
	requireOutputWithout(t, run, "soft target")
}

func TestCoverageSummaryWarnsBelowTheTargetWithoutFailing(t *testing.T) {
	t.Parallel()

	module := moduleWithCoverage(t)

	run := runScript(t, module, "", coverageSummaryScript, coverageProfile, "80")

	requirePassed(t, run)
	requireOutput(t, run, "Below the soft target of 80%.")
}

func moduleWithCoverage(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/shop\n")
	writeFile(t, dir, "internal/order/order.go", "package order\n")
	writeFile(t, dir, "internal/user/user.go", "package user\n")
	writeFile(t, dir, coverageProfile, coverageProfileContent)

	return dir
}
