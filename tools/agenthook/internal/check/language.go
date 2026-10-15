package check

import (
	"path"
	"strings"
)

const templExtension = ".templ"

func isGo(rel string) bool {
	return path.Ext(rel) == ".go"
}

func isTempl(rel string) bool {
	return path.Ext(rel) == templExtension
}

func isTypeScript(rel string) bool {
	return path.Ext(rel) == ".ts"
}

func isMigration(rel string) bool {
	return path.Dir(rel) == "migrations" && path.Ext(rel) == ".sql"
}

func generatedFromTempl(rel string) string {
	return strings.TrimSuffix(rel, templExtension) + "_templ.go"
}
