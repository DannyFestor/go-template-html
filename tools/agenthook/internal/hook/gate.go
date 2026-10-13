package hook

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/check"
	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/checkout"
	"github.com/DannyFestor/go-template-html/tools/agenthook/internal/session"
)

// After this many blocks in a row the gate lets the turn end, so an agent that cannot fix a failure hands back
// to the human instead of looping.
const maxBlocksInARow = 3

type checkoutEdits struct {
	checkout    checkout.Checkout
	editedPaths []string
	plan        check.Plan
}

type gate struct {
	store session.Store
}

func newGate(store session.Store) gate {
	return gate{store: store}
}

func (g gate) run(edits []checkoutEdits) (string, error) {
	state, err := g.store.Load()
	if err != nil {
		return "", fmt.Errorf("load the gate state: %w", err)
	}

	fingerprint := fingerprintOf(edits)

	switch {
	case state.LetThroughFingerprint == fingerprint:
		return "", nil
	case state.BlocksInARow >= maxBlocksInARow:
		return "", g.letThrough(fingerprint)
	}

	failure := firstFailure(edits)
	if failure == "" {
		return "", g.letThrough(fingerprint)
	}

	blocks := state.BlocksInARow + 1

	return blockReason(failure, blocks), g.save(func(state *session.State) { state.BlocksInARow = blocks })
}

// A release counts as letting the turn through too, so the gate stays quiet until the edits change.
func (g gate) letThrough(fingerprint string) error {
	return g.save(func(state *session.State) { state.BlocksInARow, state.LetThroughFingerprint = 0, fingerprint })
}

func (g gate) save(change func(*session.State)) error {
	err := g.store.Update(change)
	if err != nil {
		return fmt.Errorf("save the gate state: %w", err)
	}

	return nil
}

func firstFailure(edits []checkoutEdits) string {
	for _, edit := range edits {
		if failure := edit.plan.FirstFailure(check.Make{}, edit.checkout.Root); failure != "" {
			return failure
		}
	}

	return ""
}

func blockReason(failure string, blocks int) string {
	return fmt.Sprintf(
		"Fix this before ending the turn (block %d of %d in a row):\n%s",
		blocks,
		maxBlocksInARow,
		failure,
	)
}

// fingerprintOf hashes the content of every edited file, so the gate can tell nothing changed since it passed.
func fingerprintOf(edits []checkoutEdits) string {
	var listing strings.Builder

	for _, edit := range edits {
		for _, rel := range edit.editedPaths {
			listing.WriteString(
				edit.checkout.Root + "\x00" + rel + "\x00" + contentHash(edit.checkout.Root, rel) + "\n",
			)
		}
	}

	fingerprint := sha256.Sum256([]byte(listing.String()))

	return hex.EncodeToString(fingerprint[:])
}

// A deleted file hashes like an empty one, which is enough to tell it changed.
func contentHash(root, rel string) string {
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel))) //nolint:gosec // A file the agent edited.
	if err != nil {
		content = nil
	}

	hash := sha256.Sum256(content)

	return hex.EncodeToString(hash[:])
}
