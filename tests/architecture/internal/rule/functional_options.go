package rule

import (
	"go/ast"
	"slices"

	"github.com/DannyFestor/go-template-html/tests/architecture/internal/source"
)

// FunctionalOptions exists because wiring takes required dependencies positionally and adapter config as a
// struct (ADR-0002).
func FunctionalOptions(files []source.File) []Violation {
	return slices.Concat(optionTypes(files), variadicFuncParams(files))
}

func optionTypes(files []source.File) []Violation {
	var violations []Violation

	for _, file := range files {
		for _, spec := range typeSpecs(file) {
			if funcType, ok := spec.Type.(*ast.FuncType); ok && mutatesAPointer(funcType) {
				violations = append(violations, newViolation(file, spec, spec.Name.Name, "is a functional option type"))
			}
		}
	}

	return violations
}

func variadicFuncParams(files []source.File) []Violation {
	var violations []Violation

	for _, file := range files {
		for _, funcDecl := range funcDecls(file) {
			if takesVariadicFuncs(funcDecl.Type) {
				violations = append(violations,
					newViolation(file, funcDecl, funcName(funcDecl), "takes variadic function parameters"))
			}
		}
	}

	return violations
}

func mutatesAPointer(funcType *ast.FuncType) bool {
	params := funcType.Params.List
	if len(params) != 1 || len(params[0].Names) > 1 || funcType.Results != nil {
		return false
	}

	_, isPointer := params[0].Type.(*ast.StarExpr)

	return isPointer
}

func takesVariadicFuncs(funcType *ast.FuncType) bool {
	params := funcType.Params.List
	if len(params) == 0 {
		return false
	}

	ellipsis, isVariadic := params[len(params)-1].Type.(*ast.Ellipsis)
	if !isVariadic {
		return false
	}

	_, isFunc := ellipsis.Elt.(*ast.FuncType)

	return isFunc
}
