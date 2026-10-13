package hook

import (
	"fmt"

	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/check"
	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/checkout"
	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/session"
)

func (s Session) EndOfTurn(dir string) (string, error) {
	sessionCheckout, inCheckout := checkout.Find(dir)
	if !inCheckout {
		return "", nil
	}

	// The session may have edited in any worktree of the repository, not just the one it runs in.
	edits, err := s.editsIn(sessionCheckout.Worktrees())
	if err != nil || len(edits) == 0 {
		return "", err
	}

	return newGate(session.Open(sessionCheckout.GitDir, s.id)).run(edits)
}

func (s Session) editsIn(worktrees []checkout.Checkout) ([]checkoutEdits, error) {
	var edits []checkoutEdits

	for _, worktree := range worktrees {
		state, err := session.Open(worktree.GitDir, s.id).Load()
		if err != nil {
			return nil, fmt.Errorf("load the edits in %s: %w", worktree.Root, err)
		}

		plan := check.EndOfTurn(worktree.Root, state.EditedPaths)
		if len(plan) > 0 {
			edits = append(edits, checkoutEdits{checkout: worktree, editedPaths: state.EditedPaths, plan: plan})
		}
	}

	return edits, nil
}
