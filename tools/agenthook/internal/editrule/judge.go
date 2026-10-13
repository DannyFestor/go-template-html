package editrule

import (
	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/decision"
)

// Edit is one file change the agent is about to make. Path is slash-separated and relative to the checkout root.
// Removed is the text the change replaces, Inserted the text it puts in, and Original the file before it.
type Edit struct {
	Path     string
	Original string
	Removed  string
	Inserted string
}

type commentMemory interface {
	Seen(comment string) bool
	Remember(comments []string) error
}

type Judge struct {
	memory commentMemory
}

func NewJudge(memory commentMemory) Judge {
	return Judge{memory: memory}
}

// Edit runs the comment guard last, so an edit already denied leaves no comment marked as seen.
func (j Judge) Edit(edit Edit) (decision.Decision, error) {
	byPath := decision.Strongest(generatedOutput(edit), guardrailFile(edit), dependencyManifest(edit))
	if byPath.Verdict == decision.Deny {
		return byPath, nil
	}

	byComments, err := j.commentGuard(edit)

	return decision.Strongest(byPath, byComments), err
}
