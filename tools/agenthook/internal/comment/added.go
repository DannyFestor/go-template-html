package comment

import "path/filepath"

func Added(path, removed, inserted string) []string {
	syntax, known := syntaxFor(filepath.Ext(path))
	if !known {
		return nil
	}

	existing := counts(syntax.comments(removed))

	var added []string

	for _, comment := range syntax.comments(inserted) {
		switch {
		case existing[comment] > 0:
			existing[comment]--
		case !syntax.isDirective(comment):
			added = append(added, comment)
		}
	}

	return added
}

func counts(comments []string) map[string]int {
	counted := map[string]int{}
	for _, comment := range comments {
		counted[comment]++
	}

	return counted
}
