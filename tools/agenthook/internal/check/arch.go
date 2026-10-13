package check

import (
	"encoding/json"
	"slices"
	"strings"
)

// archReport is the part of go-arch-lint's JSON report that names files.
//
//nolint:tagliatelle // go-arch-lint names its JSON fields in PascalCase.
type archReport struct {
	Payload struct {
		ArchWarningsDeps       []archWarning `json:"ArchWarningsDeps"`
		ArchWarningsNotMatched []archWarning `json:"ArchWarningsNotMatched"`
	} `json:"Payload"`
}

//nolint:tagliatelle // go-arch-lint names its JSON fields in PascalCase.
type archWarning struct {
	ComponentName      string `json:"ComponentName"`
	FileRelativePath   string `json:"FileRelativePath"`
	ResolvedImportName string `json:"ResolvedImportName"`
}

// A warning without an import is a file outside every component.
func (w archWarning) String() string {
	if w.ResolvedImportName == "" {
		return w.file() + ": no component in .go-arch-lint.yml holds it"
	}

	return w.file() + ": component " + w.ComponentName + " may not import " + w.ResolvedImportName
}

func (w archWarning) file() string {
	return strings.TrimPrefix(w.FileRelativePath, "/")
}

// archCheck fails only on warnings about files, relative to the checkout root, that the session edited.
type archCheck struct {
	files []string
}

func (a archCheck) run(maker runner, root string) string {
	run := maker.Run(root, checkoutTarget("arch-json", nil))
	if !run.Failed {
		return ""
	}

	var report archReport

	err := json.Unmarshal([]byte(run.Stdout), &report)
	if err != nil {
		return failure(run)
	}

	return a.describe(report)
}

func (a archCheck) describe(report archReport) string {
	var lines []string

	for _, warning := range append(report.Payload.ArchWarningsDeps, report.Payload.ArchWarningsNotMatched...) {
		if slices.Contains(a.files, warning.file()) {
			lines = append(lines, warning.String())
		}
	}

	if len(lines) == 0 {
		return ""
	}

	return "`make arch` failed:\n" + strings.Join(lines, "\n")
}
