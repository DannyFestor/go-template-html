package asset

import "io/fs"

type fixedManifest struct {
	manifest manifest
}

func (f fixedManifest) load() (manifest, error) {
	return f.manifest, nil
}

// liveManifest reads the manifest on every lookup, because each rebuild of `make watch-vite` changes the hashes.
type liveManifest struct {
	files fs.FS
}

func (l liveManifest) load() (manifest, error) {
	return readManifest(l.files)
}
