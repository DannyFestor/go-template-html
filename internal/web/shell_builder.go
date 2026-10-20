package web

import (
	"context"
	"fmt"

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

type ShellBuilder struct {
	appName string
	assets  assetURLs
	flashes flashTaker
	users   currentUser
}

func NewShellBuilder(appName string, assets assetURLs, flashes flashTaker, users currentUser) *ShellBuilder {
	return &ShellBuilder{appName: appName, assets: assets, flashes: flashes, users: users}
}

func (b *ShellBuilder) Build(ctx context.Context) (view.Shell, error) {
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
		User:        b.shellUser(ctx),
		Status:      b.flashes.TakeFlash(ctx),
	}, nil
}

func (b *ShellBuilder) shellUser(ctx context.Context) view.ShellUser {
	user, signedIn := b.users.CurrentUser(ctx)
	if !signedIn {
		return view.ShellUser{Name: "", Email: "", Initials: ""}
	}

	return view.ShellUser{Name: user.Name, Email: user.Email, Initials: initials(user.Name)}
}
