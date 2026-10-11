package rule

import (
	"go/ast"
	"maps"
	"slices"
	"strings"

	"github.com/DannyFestor/go-template-html/tests/architecture/internal/source"
)

type declaredType struct {
	file source.File
	spec *ast.TypeSpec
}

type packageIndex struct {
	types   map[string]declaredType
	funcs   map[string]*ast.FuncDecl
	methods map[string][]*ast.FuncDecl
}

func indexPackages(files []source.File) []packageIndex {
	byDir := map[string][]source.File{}
	for _, file := range files {
		byDir[file.Dir()] = append(byDir[file.Dir()], file)
	}

	indexes := make([]packageIndex, 0, len(byDir))
	for _, dir := range slices.Sorted(maps.Keys(byDir)) {
		indexes = append(indexes, indexPackage(byDir[dir]))
	}

	return indexes
}

func indexPackage(files []source.File) packageIndex {
	index := packageIndex{
		types:   map[string]declaredType{},
		funcs:   map[string]*ast.FuncDecl{},
		methods: map[string][]*ast.FuncDecl{},
	}

	for _, file := range files {
		index.addTypes(file)
		index.addFuncs(file)
	}

	return index
}

func (p packageIndex) addTypes(file source.File) {
	for _, spec := range typeSpecs(file) {
		p.types[spec.Name.Name] = declaredType{file: file, spec: spec}
	}
}

func (p packageIndex) addFuncs(file source.File) {
	for _, funcDecl := range funcDecls(file) {
		receiver, isMethod := receiverOf(funcDecl)
		if !isMethod {
			p.funcs[funcDecl.Name.Name] = funcDecl

			continue
		}

		p.methods[receiver] = append(p.methods[receiver], funcDecl)
	}
}

func (p packageIndex) isStruct(name string) bool {
	declared, ok := p.types[name]
	if !ok {
		return false
	}

	_, isStruct := declared.spec.Type.(*ast.StructType)

	return isStruct
}

// isCompanion tells <X>In and <X>Out apart from an action whose own name ends in In or Out, such as SignIn.
func (p packageIndex) isCompanion(name string) bool {
	for _, suffix := range []string{"In", "Out"} {
		if !strings.HasSuffix(name, suffix) {
			continue
		}

		if _, hasAction := p.types[strings.TrimSuffix(name, suffix)]; hasAction {
			return true
		}
	}

	return false
}

func (p packageIndex) exportedMethods(typeName string) []*ast.FuncDecl {
	return slices.DeleteFunc(slices.Clone(p.methods[typeName]), func(method *ast.FuncDecl) bool {
		return !method.Name.IsExported()
	})
}
