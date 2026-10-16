package rule

import (
	"go/ast"

	"github.com/DannyFestor/go-template-html/tests/architecture/internal/source"
)

func GoStatements(files []source.File) []Violation {
	return funcsContaining(files, isGoStatement, "starts a goroutine")
}

func isGoStatement(node ast.Node) bool {
	_, isGo := node.(*ast.GoStmt)

	return isGo
}
