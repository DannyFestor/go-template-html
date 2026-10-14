package asset

import (
	"path"
	"strings"
)

const viteHashSeparator = "-"

func withoutViteHash(file string) (string, bool) {
	extension := path.Ext(file)
	stem := strings.TrimSuffix(file, extension)

	hashStart := strings.LastIndex(stem, viteHashSeparator)
	if hashStart <= strings.LastIndex(stem, "/") {
		return "", false
	}

	return stem[:hashStart] + extension, true
}

func sameViteName(file, other string) bool {
	name, hashed := withoutViteHash(file)
	otherName, otherHashed := withoutViteHash(other)

	return hashed && otherHashed && name == otherName
}
