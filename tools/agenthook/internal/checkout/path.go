package checkout

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// realPath resolves the symlinks in the part of path that exists, because git reports resolved paths, as macOS
// does for /var under /private/var.
func realPath(path string) string {
	existing := nearestExisting(path)

	resolved, err := filepath.EvalSymlinks(existing)
	if err != nil {
		return path
	}

	rest, err := filepath.Rel(existing, path)
	if err != nil {
		return path
	}

	return filepath.Join(resolved, rest)
}

func nearestExisting(path string) string {
	for {
		_, err := os.Lstat(path)
		if !errors.Is(err, fs.ErrNotExist) || filepath.Dir(path) == path {
			return path
		}

		path = filepath.Dir(path)
	}
}

func nearestExistingDir(path string) string {
	existing := nearestExisting(path)

	info, err := os.Stat(existing)
	if err != nil || !info.IsDir() {
		return filepath.Dir(existing)
	}

	return existing
}
