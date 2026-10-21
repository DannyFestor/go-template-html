package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/view"
)

const appEntry = "internal/asset/src/app.ts"

type assetURLs interface {
	ScriptURL(source string) (string, error)
	StylesheetURLs(source string) ([]string, error)
}

type flashTaker interface {
	TakeFlash(ctx context.Context) string
}

type currentUser interface {
	CurrentUser(ctx context.Context) (domain.User, bool)
}

type themeReader interface {
	Read(r *http.Request) domain.Theme
}

type ShellBuilder struct {
	appName string
	assets  assetURLs
	flashes flashTaker
	users   currentUser
	themes  themeReader
}

func NewShellBuilder(
	appName string,
	assets assetURLs,
	flashes flashTaker,
	users currentUser,
	themes themeReader,
) *ShellBuilder {
	return &ShellBuilder{appName: appName, assets: assets, flashes: flashes, users: users, themes: themes}
}

func (b *ShellBuilder) Build(req *http.Request) (view.Shell, error) {
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
		User:        b.shellUser(req.Context()),
		Status:      b.flashes.TakeFlash(req.Context()),
		Theme:       string(b.themes.Read(req)),
	}, nil
}

func (b *ShellBuilder) shellUser(ctx context.Context) view.ShellUser {
	user, signedIn := b.users.CurrentUser(ctx)
	if !signedIn {
		return view.ShellUser{ID: "", Name: "", Email: "", Initials: ""}
	}

	return view.ShellUser{ID: string(user.ID), Name: user.Name, Email: user.Email, Initials: initials(user.Name)}
}
