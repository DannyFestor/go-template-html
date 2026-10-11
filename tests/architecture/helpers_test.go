package architecture_test

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/DannyFestor/go-template-html/tests/architecture/internal/rule"
	"github.com/DannyFestor/go-template-html/tests/architecture/internal/source"
)

const moduleRoot = "../.."

func loadModule(t *testing.T) []source.File {
	t.Helper()

	return load(t, moduleRoot)
}

func loadFixture(t *testing.T, name string) []source.File {
	t.Helper()

	return load(t, filepath.Join("testdata", name))
}

func load(t *testing.T, root string) []source.File {
	t.Helper()

	files, err := source.Load(root)
	if err != nil {
		t.Fatalf("load %s: %v", root, err)
	}

	return files
}

// Each allowlist entry maps a violation subject to the reason it is accepted. An entry that matches no
// violation fails too, so the list cannot outlive its exceptions.
func requireNoViolations(t *testing.T, violations []rule.Violation, allowlist map[string]string) {
	t.Helper()

	used := map[string]bool{}

	for _, violation := range violations {
		if _, allowed := allowlist[violation.Subject]; allowed {
			used[violation.Subject] = true

			continue
		}

		t.Error(violation)
	}

	for subject := range allowlist {
		if !used[subject] {
			t.Errorf("allowlist entry %q matches no violation", subject)
		}
	}
}

func requireSubjects(t *testing.T, violations []rule.Violation, want ...string) {
	t.Helper()

	got := make([]string, 0, len(violations))
	for _, violation := range violations {
		got = append(got, violation.Subject)
	}

	slices.Sort(got)
	slices.Sort(want)

	if !slices.Equal(got, want) {
		t.Errorf("violation subjects = %q, want %q", got, want)
	}
}
