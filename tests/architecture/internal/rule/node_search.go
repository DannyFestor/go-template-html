package rule

import (
	"go/ast"

	"github.com/DannyFestor/go-template-html/tests/architecture/internal/source"
)

func funcsContaining(files []source.File, matches func(ast.Node) bool, message string) []Violation {
	var violations []Violation

	for _, file := range files {
		for _, funcDecl := range funcDecls(file) {
			if containsNode(funcDecl, matches) {
				violations = append(violations, newViolation(file, funcDecl, funcName(funcDecl), message))
			}
		}
	}

	return violations
}

func containsNode(decl *ast.FuncDecl, matches func(ast.Node) bool) bool {
	found := false

	ast.Inspect(decl, func(node ast.Node) bool {
		found = found || matches(node)

		return !found
	})

	return found
}
