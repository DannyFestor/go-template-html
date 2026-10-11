package rule

import (
	"fmt"

	"github.com/DannyFestor/go-template-html/tests/architecture/internal/source"
)

// Violation is one broken rule. Subject names the offending declaration and is the key allowlists use.
type Violation struct {
	Subject string
	Path    string
	Line    int
	Message string
}

func (v Violation) String() string {
	return fmt.Sprintf("%s:%d: %s: %s", v.Path, v.Line, v.Subject, v.Message)
}

func newViolation(file source.File, node positioned, subject, message string) Violation {
	return Violation{
		Subject: file.Dir() + "." + subject,
		Path:    file.Path,
		Line:    file.Line(node.Pos()),
		Message: message,
	}
}
