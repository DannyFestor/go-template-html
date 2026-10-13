package shell

import "path/filepath"

func Resolve(dir, path string) string {
	switch {
	// "-" is cd's previous directory, which the parser does not track.
	case path == "" || path == "-":
		return ""
	case filepath.IsAbs(path):
		return filepath.Clean(path)
	case dir == "":
		return ""
	default:
		return filepath.Join(dir, path)
	}
}
