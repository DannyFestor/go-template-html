package rule

import (
	"go/ast"

	"github.com/DannyFestor/go-template-html/tests/architecture/internal/source"
)

// ExportedInterfaces counts type constraints too, because a consumer declares its own like any other interface.
func ExportedInterfaces(files []source.File) []Violation {
	var violations []Violation

	for _, file := range files {
		for _, spec := range typeSpecs(file) {
			if _, isInterface := spec.Type.(*ast.InterfaceType); isInterface && spec.Name.IsExported() {
				violations = append(violations, newViolation(file, spec, spec.Name.Name, "is an exported interface"))
			}
		}
	}

	return violations
}
