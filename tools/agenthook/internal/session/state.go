package session

import "slices"

// State is what one agent session remembers in one checkout. The last two fields belong to the end-of-turn gate.
type State struct {
	EditedPaths           []string `json:"editedPaths"`
	SeenComments          []string `json:"seenComments"`
	BlocksInARow          int      `json:"blocksInARow"`
	LetThroughFingerprint string   `json:"letThroughFingerprint"`
}

func emptyState() State {
	return State{EditedPaths: nil, SeenComments: nil, BlocksInARow: 0, LetThroughFingerprint: ""}
}

func appendNew(values []string, added ...string) []string {
	for _, value := range added {
		if !slices.Contains(values, value) {
			values = append(values, value)
		}
	}

	return values
}
