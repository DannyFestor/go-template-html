//go:build feature

package feature_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	builtManifest = "../../internal/asset/dist/manifest.json"
	appEntry      = "internal/asset/src/app.ts"
)

type builtChunk struct {
	File string   `json:"file"`
	CSS  []string `json:"css"`
}

func builtAppEntry(t *testing.T) builtChunk {
	t.Helper()

	content, err := os.ReadFile(builtManifest)
	require.NoError(t, err, "run `make vite-build` before the feature tests")

	var manifest map[string]builtChunk

	require.NoError(t, json.Unmarshal(content, &manifest))
	require.Contains(t, manifest, appEntry)

	return manifest[appEntry]
}

func prefixed(prefix string, values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, prefix+value)
	}

	return result
}
