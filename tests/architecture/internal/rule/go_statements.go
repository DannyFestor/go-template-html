package rule

import (
	"go/ast"

	"github.com/DannyFestor/go-template-html/tests/architecture/internal/source"
)

func GoStatements(files []source.File) []Violation {
	var violations []Violation

	for _, file := range files {
		for _, funcDecl := range funcDecls(file) {
			if startsGoroutine(funcDecl) {
				violations = append(violations, newViolation(file, funcDecl, funcName(funcDecl), "starts a goroutine"))
			}
		}
	}

	return violations
}

func startsGoroutine(decl *ast.FuncDecl) bool {
	found := false

	ast.Inspect(decl, func(node ast.Node) bool {
		_, isGo := node.(*ast.GoStmt)
		found = found || isGo

		return !found
	})

	return found
}
