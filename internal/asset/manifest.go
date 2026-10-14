package asset

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"slices"
)

const manifestPath = "manifest.json"

var errUnknownEntry = errors.New("no entry in the asset manifest")

type chunk struct {
	File string   `json:"file"`
	CSS  []string `json:"css"`
}

type manifest map[string]chunk

type resolved struct {
	file    string
	current bool
}

func readManifest(files fs.FS) (manifest, error) {
	content, err := fs.ReadFile(files, manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read the asset manifest: %w", err)
	}

	var parsed manifest

	err = json.Unmarshal(content, &parsed)
	if err != nil {
		return nil, fmt.Errorf("parse the asset manifest: %w", err)
	}

	return parsed, nil
}

// resolve also matches a file from an older build by its name, because HTML rendered before a deploy still links it.
func (m manifest) resolve(requested string) (resolved, bool) {
	emitted := m.emitted()

	if slices.Contains(emitted, requested) {
		return resolved{file: requested, current: true}, true
	}

	for _, file := range emitted {
		if sameViteName(file, requested) {
			return resolved{file: file, current: false}, true
		}
	}

	return resolved{file: "", current: false}, false
}

func (m manifest) emitted() []string {
	var files []string

	for _, emitted := range m {
		files = append(files, emitted.File)
		files = append(files, emitted.CSS...)
	}

	return files
}

func (m manifest) entry(source string) (chunk, error) {
	found, ok := m[source]
	if !ok {
		return chunk{}, fmt.Errorf("%w: %q", errUnknownEntry, source)
	}

	return found, nil
}
