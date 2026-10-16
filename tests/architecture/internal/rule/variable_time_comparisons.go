package rule

import (
	"go/ast"
	"go/token"

	"github.com/DannyFestor/go-template-html/tests/architecture/internal/source"
)

func VariableTimeComparisons(files []source.File) []Violation {
	return funcsContaining(files, isVariableTimeComparison,
		"compares bytes in variable time, use hmac.Equal or crypto/subtle")
}

func isVariableTimeComparison(node ast.Node) bool {
	return isVariableTimeCall(node) || isStringConversionComparison(node)
}

func isVariableTimeCall(node ast.Node) bool {
	call, ok := node.(*ast.CallExpr)

	return ok && (isSelector(call.Fun, "bytes", "Equal") || isSelector(call.Fun, "bytes", "Compare"))
}

func isStringConversionComparison(node ast.Node) bool {
	binary, ok := node.(*ast.BinaryExpr)
	if !ok || (binary.Op != token.EQL && binary.Op != token.NEQ) {
		return false
	}

	return isStringConversion(binary.X) || isStringConversion(binary.Y)
}

func isStringConversion(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)

	return ok && len(call.Args) == 1 && isIdent(call.Fun, "string")
}
