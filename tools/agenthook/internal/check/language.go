package check

import "path"

func isGo(rel string) bool {
	return path.Ext(rel) == ".go"
}

func isTypeScript(rel string) bool {
	return path.Ext(rel) == ".ts"
}
