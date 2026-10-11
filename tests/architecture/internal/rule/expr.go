package rule

import "go/ast"

// fieldTypes lists one type per declared name, so `a, b T` yields T twice.
func fieldTypes(fields *ast.FieldList) []ast.Expr {
	if fields == nil {
		return nil
	}

	var types []ast.Expr

	for _, field := range fields.List {
		for range max(1, len(field.Names)) {
			types = append(types, field.Type)
		}
	}

	return types
}

func isIdent(expr ast.Expr, name string) bool {
	ident, ok := expr.(*ast.Ident)

	return ok && ident.Name == name
}

func isPointerTo(expr ast.Expr, name string) bool {
	star, ok := expr.(*ast.StarExpr)

	return ok && isIdent(star.X, name)
}

func isSelector(expr ast.Expr, pkg, name string) bool {
	selector, ok := expr.(*ast.SelectorExpr)

	return ok && isIdent(selector.X, pkg) && selector.Sel.Name == name
}
