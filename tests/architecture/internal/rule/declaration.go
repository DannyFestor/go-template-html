package rule

import (
	"go/ast"
	"go/token"

	"github.com/DannyFestor/go-template-html/tests/architecture/internal/source"
)

type positioned interface {
	Pos() token.Pos
}

func funcName(decl *ast.FuncDecl) string {
	receiver, isMethod := receiverOf(decl)
	if !isMethod {
		return decl.Name.Name
	}

	return receiver + "." + decl.Name.Name
}

func receiverOf(decl *ast.FuncDecl) (string, bool) {
	if decl.Recv == nil || len(decl.Recv.List) == 0 {
		return "", false
	}

	return receiverTypeName(decl.Recv.List[0].Type), true
}

func funcDecls(file source.File) []*ast.FuncDecl {
	var decls []*ast.FuncDecl

	for _, decl := range file.Syntax.Decls {
		if funcDecl, isFunc := decl.(*ast.FuncDecl); isFunc {
			decls = append(decls, funcDecl)
		}
	}

	return decls
}

func receiverTypeName(expr ast.Expr) string {
	switch typed := expr.(type) {
	case *ast.StarExpr:
		return receiverTypeName(typed.X)
	case *ast.IndexExpr:
		return receiverTypeName(typed.X)
	case *ast.IndexListExpr:
		return receiverTypeName(typed.X)
	case *ast.Ident:
		return typed.Name
	default:
		return ""
	}
}

func typeSpecs(file source.File) []*ast.TypeSpec {
	var specs []*ast.TypeSpec

	for _, decl := range file.Syntax.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}

		for _, spec := range genDecl.Specs {
			if typeSpec, isType := spec.(*ast.TypeSpec); isType {
				specs = append(specs, typeSpec)
			}
		}
	}

	return specs
}
