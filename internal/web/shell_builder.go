package web

import (
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/view"
)

const appEntry = "internal/asset/src/app.ts"

type assetURLs interface {
	ScriptURL(source string) (string, error)
	StylesheetURLs(source string) ([]string, error)
}

type ShellBuilder struct {
	appName string
	assets  assetURLs
}

func NewShellBuilder(appName string, assets assetURLs) *ShellBuilder {
	return &ShellBuilder{appName: appName, assets: assets}
}

func (b *ShellBuilder) Build() (view.Shell, error) {
	script, err := b.assets.ScriptURL(appEntry)
	if err != nil {
		return view.Shell{}, fmt.Errorf("resolve the app script: %w", err)
	}

	stylesheets, err := b.assets.StylesheetURLs(appEntry)
	if err != nil {
		return view.Shell{}, fmt.Errorf("resolve the app stylesheets: %w", err)
	}

	return view.Shell{
		AppName:     b.appName,
		Script:      script,
		Stylesheets: stylesheets,
		User:        view.ShellUser{Name: "", Email: "", Initials: ""},
	}, nil
}
