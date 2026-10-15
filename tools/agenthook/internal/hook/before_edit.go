package hook

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/checkout"
	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/decision"
	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/editrule"
	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/session"
)

// FileEdit is a file change the agent is about to make. A WholeFile change replaces the file, so the text it
// removes is the file on disk.
type FileEdit struct {
	Path      string
	Removed   string
	Inserted  string
	WholeFile bool
}

func (s Session) BeforeEdit(edit FileEdit) (decision.Decision, error) {
	found, rel, inCheckout := checkout.Locate(edit.Path)
	if !inCheckout {
		return decision.Passed(), nil
	}

	original, exists, err := readOptionalFile(edit.Path)
	if err != nil {
		return decision.Passed(), err
	}

	removed := edit.Removed
	if edit.WholeFile {
		removed = original
	}

	judge := editrule.NewJudge(session.Open(found.GitDir, s.id), found)

	judged, err := judge.Edit(editrule.Edit{
		Path:     rel,
		Original: original,
		Removed:  removed,
		Inserted: edit.Inserted,
		Created:  !exists,
	})
	if err != nil {
		return judged, fmt.Errorf("remember the comments: %w", err)
	}

	return judged, nil
}

func readOptionalFile(path string) (string, bool, error) {
	content, err := os.ReadFile(path) //nolint:gosec // Reading the file the agent is about to edit is the point.
	if errors.Is(err, fs.ErrNotExist) {
		return "", false, nil
	}

	if err != nil {
		return "", false, fmt.Errorf("read the file before the edit: %w", err)
	}

	return string(content), true, nil
}
