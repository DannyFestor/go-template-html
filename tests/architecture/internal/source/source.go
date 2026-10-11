package source

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

// File is a hand-written, non-test Go file. Path is slash-separated and relative to the loaded root.
type File struct {
	Path   string
	Fset   *token.FileSet
	Syntax *ast.File
}

func (f File) Dir() string {
	return path.Dir(f.Path)
}

func (f File) Line(pos token.Pos) int {
	return f.Fset.Position(pos).Line
}

// Load skips tests and generated code, because the rules govern hand-written production code.
func Load(root string) ([]File, error) {
	fset := token.NewFileSet()

	var files []File

	err := filepath.WalkDir(root, func(filePath string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return skipIgnoredDir(filePath, root, entry.Name())
		}

		if !isNonTestGoFile(filePath) {
			return nil
		}

		file, err := parse(fset, root, filePath)
		if err == nil && !ast.IsGenerated(file.Syntax) {
			files = append(files, file)
		}

		return err
	})
	if err != nil {
		return nil, fmt.Errorf("load %s: %w", root, err)
	}

	return files, nil
}

func Under(files []File, dirs ...string) []File {
	return slices.DeleteFunc(slices.Clone(files), func(file File) bool {
		return !slices.ContainsFunc(dirs, func(dir string) bool {
			return strings.HasPrefix(file.Path, dir+"/")
		})
	})
}

// The go command ignores testdata and dot or underscore directories; vendor and node_modules hold others' code.
func skipIgnoredDir(dirPath, root, name string) error {
	if dirPath == root {
		return nil
	}

	ignored := name == "testdata" || name == "vendor" || name == "node_modules" ||
		strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
	if ignored {
		return filepath.SkipDir
	}

	return nil
}

func isNonTestGoFile(filePath string) bool {
	return filepath.Ext(filePath) == ".go" && !strings.HasSuffix(filePath, "_test.go")
}

func parse(fset *token.FileSet, root, filePath string) (File, error) {
	syntax, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return File{}, fmt.Errorf("parse: %w", err)
	}

	relative, err := filepath.Rel(root, filePath)
	if err != nil {
		return File{}, fmt.Errorf("relative path: %w", err)
	}

	return File{Path: filepath.ToSlash(relative), Fset: fset, Syntax: syntax}, nil
}
