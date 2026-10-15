package checkout

import (
	"path/filepath"
	"strings"
)

const mainlineRef = "origin/main"

// Checkout is one git work tree, the main checkout or a linked worktree. GitDir is its own git dir, which a
// linked worktree does not share.
type Checkout struct {
	Root   string
	GitDir string
}

// Find returns the checkout holding path, which need not exist yet. A relative path would be judged against the
// hook's own working directory, so it has no checkout.
func Find(path string) (Checkout, bool) {
	if !filepath.IsAbs(path) {
		return Checkout{Root: "", GitDir: ""}, false
	}

	output, err := gitOutput(nearestExistingDir(realPath(path)), "rev-parse", "--show-toplevel", "--absolute-git-dir")
	if err != nil {
		return Checkout{Root: "", GitDir: ""}, false
	}

	root, gitDir, _ := strings.Cut(output, "\n")

	return Checkout{Root: root, GitDir: gitDir}, true
}

func Locate(path string) (Checkout, string, bool) {
	found, inCheckout := Find(path)
	rel, inside := found.Rel(path)

	return found, rel, inCheckout && inside
}

func (c Checkout) Rel(path string) (string, bool) {
	rel, err := filepath.Rel(c.Root, realPath(path))
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}

	return filepath.ToSlash(rel), true
}

func (c Checkout) Branch() string {
	branch, err := gitOutput(c.Root, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return ""
	}

	return branch
}

func (c Checkout) OnMainline(rel string) bool {
	_, err := gitOutput(c.Root, "cat-file", "-e", mainlineRef+":"+rel)

	return err == nil
}

func (c Checkout) Worktrees() []Checkout {
	output, err := gitOutput(c.Root, "worktree", "list", "--porcelain")
	if err != nil {
		return []Checkout{c}
	}

	var worktrees []Checkout

	for line := range strings.Lines(output) {
		path, isWorktree := strings.CutPrefix(strings.TrimSpace(line), "worktree ")
		if worktree, found := Find(path); isWorktree && found {
			worktrees = append(worktrees, worktree)
		}
	}

	return worktrees
}
