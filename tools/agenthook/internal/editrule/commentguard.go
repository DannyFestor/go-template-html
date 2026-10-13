package editrule

import (
	"strings"

	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/comment"
	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/decision"
)

const (
	addedCommentsReason = "Added comments are refused: comments only explain a why the code cannot. Remove them, " +
		"or resubmit the same comments to ask the human if each one explains such a why:"
	resubmittedCommentsReason = "The agent resubmitted comments it holds to explain a why the code cannot:"
)

// The first submission of a comment is denied and the same comment resubmitted asks the human, so the agent has
// to reconsider each comment once.
func (j Judge) commentGuard(edit Edit) (decision.Decision, error) {
	added := comment.Added(edit.Path, edit.Removed, edit.Inserted)

	switch {
	case len(added) == 0:
		return decision.Passed(), nil
	case j.allSeen(added):
		return decision.Asked(listed(resubmittedCommentsReason, added)), nil
	default:
		return decision.Denied(listed(addedCommentsReason, added)), j.memory.Remember(added)
	}
}

func (j Judge) allSeen(comments []string) bool {
	for _, comment := range comments {
		if !j.memory.Seen(comment) {
			return false
		}
	}

	return true
}

func listed(reason string, comments []string) string {
	return reason + "\n  " + strings.Join(comments, "\n  ")
}
