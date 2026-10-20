package request

import (
	"net/url"
	"strings"
)

func Field(form url.Values, name string) string {
	return strings.TrimSpace(form.Get(name))
}

func SecretField(form url.Values, name string) string {
	return form.Get(name)
}
