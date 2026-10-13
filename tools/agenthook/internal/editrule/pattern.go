package editrule

import (
	"path"
	"slices"
	"strings"
)

// A pattern ending in "/**" matches everything below its directory, and one starting with "**/" matches its file
// name in any directory. Any other pattern is a path.Match pattern for the whole path.
func matches(pattern, filePath string) bool {
	switch {
	case strings.HasSuffix(pattern, "/**"):
		return strings.HasPrefix(filePath, strings.TrimSuffix(pattern, "**"))
	case strings.HasPrefix(pattern, "**/"):
		return path.Base(filePath) == strings.TrimPrefix(pattern, "**/")
	default:
		matched, err := path.Match(pattern, filePath)

		return err == nil && matched
	}
}

func matchesAny(patterns []string, filePath string) bool {
	return slices.ContainsFunc(patterns, func(pattern string) bool { return matches(pattern, filePath) })
}
