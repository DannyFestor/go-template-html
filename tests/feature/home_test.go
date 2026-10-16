//go:build feature

package feature_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestHomeRendersInsideThePublicLayout(t *testing.T) {
	t.Parallel()

	document := startApp(t).page(t, "/")

	main := document.Find("body > main#public-main")
	assert.Equal(t, appName, document.Find("head > title").Text())
	assert.Equal(t, "Let's get started", main.Find("h1").Text())
	assert.Equal(t, 1, main.Find(`header a[href="/login"]`).Length())
	assert.Equal(t, 1, main.Find(`header a[href="/register"]`).Length())
}

func TestHomeLinksTheHashedAssetsFromTheManifest(t *testing.T) {
	t.Parallel()

	built := builtAppEntry(t)

	document := startApp(t).page(t, "/")

	script, _ := document.Find(`head > script[type="module"]`).Attr("src")
	assert.Equal(t, "/"+built.File, script)

	var stylesheets []string

	document.Find(`head > link[rel="stylesheet"]`).Each(func(_ int, link *goquery.Selection) {
		stylesheets = append(stylesheets, link.AttrOr("href", ""))
	})
	assert.Equal(t, prefixed("/", built.CSS), stylesheets)
}

func TestHomeHasNoInlineScriptOrStyle(t *testing.T) {
	t.Parallel()

	document := startApp(t).page(t, "/")

	document.Find("script").Each(func(_ int, script *goquery.Selection) {
		assert.NotEmpty(t, script.AttrOr("src", ""), "a script without src")
		assert.Empty(t, script.Text(), "a script with inline code")
	})
	assert.Zero(t, document.Find("style").Length(), "a <style> element")
	assert.Empty(t, inlineAttributes(document), "inline style or event handler attributes")
}

func TestProductionServesTheAppScriptForAYear(t *testing.T) {
	t.Parallel()

	fetched := startApp(t).get(t, "/"+builtAppEntry(t).File)

	assert.Equal(t, http.StatusOK, fetched.status)
	assert.Equal(t, "application/javascript", fetched.header.Get("Content-Type"))
	assert.Equal(t, "public, max-age=31536000, immutable", fetched.header.Get("Cache-Control"))
}

func inlineAttributes(document *goquery.Document) []string {
	var found []string

	document.Find("*").Each(func(_ int, element *goquery.Selection) {
		for _, attribute := range element.Nodes[0].Attr {
			if attribute.Key == "style" || strings.HasPrefix(attribute.Key, "on") {
				found = append(found, goquery.NodeName(element)+"["+attribute.Key+"]")
			}
		}
	})

	return found
}
