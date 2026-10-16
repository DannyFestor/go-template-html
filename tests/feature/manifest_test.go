//go:build feature

package feature_test

import (
	"encoding/json"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	builtDist     = "../../internal/asset/dist/"
	builtManifest = builtDist + "manifest.json"
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

func builtFile(t *testing.T, name string) []byte {
	t.Helper()

	content, err := fs.ReadFile(os.DirFS(builtDist), name)
	require.NoError(t, err, "run `make vite-build` before the feature tests")

	return content
}

func prefixed(prefix string, values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, prefix+value)
	}

	return result
}
