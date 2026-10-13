package check

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

// goModule returns the directory of the go.mod nearest to rel, relative to the checkout root.
func goModule(root, rel string) string {
	for dir := path.Dir(rel); ; dir = path.Dir(dir) {
		_, err := os.Stat(filepath.Join(root, filepath.FromSlash(dir), "go.mod"))
		if err == nil || dir == "." {
			return dir
		}
	}
}

func moduleRel(module, rel string) string {
	if module == "." {
		return rel
	}

	return strings.TrimPrefix(rel, module+"/")
}
