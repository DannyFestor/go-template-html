// PROTOTYPE, throwaway. Answers one question for the "Pages, layout and component set" ticket:
// what do the app shell, auth pages and settings pages look like?
//
// Three variants, switchable from the floating bar (or ←/→), persisted in a cookie.
// The page bodies (forms, copy) are shared; only the shells and the settings frame differ:
//
//	A  Livewire default: sidebar app layout, settings sub-nav inside <main>, simple centred auth pages
//	B  Header app layout, settings as horizontal tabs, auth pages in a card
//	C  Sidebar with settings nested in it, wide two-column settings sections, split-screen auth pages
//
// Every route of the template exists. No real auth: /dashboard and /settings/* render the app
// layout, everything else the public layout. Forms are stubs that redirect back with a status.
// Navigation is boosted per ADR-0004 (only each layout's <main> swaps, server retargets crossings).
//
// Run: go run ./prototype/pages/main.go  →  http://localhost:8098
package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"
)

var variants = []struct{ Key, Name string }{
	{"A", "Sidebar · sub-nav settings · simple auth"},
	{"B", "Header · tabbed settings · card auth"},
	{"C", "Sidebar with nested settings · two-column settings · split auth"},
}

type route struct {
	Title, Body string
	App         bool
	Settings    bool
}

var routes = map[string]route{
	"/":                     {Title: "Welcome", Body: "home"},
	"/login":                {Title: "Log in", Body: "login"},
	"/register":             {Title: "Register", Body: "register"},
	"/forgot-password":      {Title: "Forgot password", Body: "forgot-password"},
	"/reset-password":       {Title: "Reset password", Body: "reset-password"},
	"/verify-email":         {Title: "Verify email", Body: "verify-email"},
	"/confirm-password":     {Title: "Confirm password", Body: "confirm-password"},
	"/two-factor-challenge": {Title: "Two-factor authentication", Body: "two-factor-challenge"},
	"/dashboard":            {Title: "Dashboard", Body: "dashboard", App: true},
	"/settings/profile":     {Title: "Profile", Body: "settings-profile", App: true, Settings: true},
	"/settings/password":    {Title: "Password", Body: "settings-password", App: true, Settings: true},
	"/settings/two-factor":  {Title: "Two-factor authentication", Body: "settings-two-factor", App: true, Settings: true},
	"/settings/appearance":  {Title: "Appearance", Body: "settings-appearance", App: true, Settings: true},
}

type page struct {
	route
	Path, Variant, VariantName, Theme, Status string
	TwoFactor                                 bool
	RecoveryCodes                             []string
	UserName, UserEmail, Initials             string
}

var twoFactorEnabled = false

func main() {
	tpl := template.Must(template.New("").Funcs(template.FuncMap{
		"dict": func(kv ...any) map[string]any {
			m := map[string]any{}
			for i := 0; i+1 < len(kv); i += 2 {
				m[kv[i].(string)] = kv[i+1]
			}
			return m
		},
	}).Parse(templates))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rt, ok := routes[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		if handleStub(w, r) {
			return
		}
		p := page{
			route: rt, Path: r.URL.Path, Variant: variant(w, r), Theme: theme(r),
			Status: r.URL.Query().Get("status"), TwoFactor: twoFactorEnabled,
			RecoveryCodes: []string{"7c1f-29ab-e04d", "b82e-5f1c-90a7", "31d9-c6e2-7b4f", "e5a0-8d37-12fc", "4f6b-a91e-d802", "9c27-3be5-6a1d", "d0f8-74c9-e3b6"},
			UserName:      "Danny Festor", UserEmail: "danny@example.com", Initials: "DF",
		}
		for _, v := range variants {
			if v.Key == p.Variant {
				p.VariantName = v.Name
			}
		}
		retargetLayoutCrossing(w, r, rt.App)
		if err := tpl.ExecuteTemplate(w, layoutName(p), p); err != nil {
			log.Println(err)
		}
	})
	log.Println("prototype on http://localhost:8098")
	log.Fatal(http.ListenAndServe(":8098", nil))
}

// handleStub fakes the POST, redirect, GET side of every form so flows can be clicked through.
func handleStub(w http.ResponseWriter, r *http.Request) bool {
	q := r.URL.Query()
	switch {
	case q.Has("theme"):
		http.SetCookie(w, &http.Cookie{Name: "theme", Value: q.Get("theme"), Path: "/"})
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
	case q.Has("two-factor"):
		twoFactorEnabled = q.Get("two-factor") == "on"
		http.Redirect(w, r, r.URL.Path+"?status=two-factor-"+q.Get("two-factor"), http.StatusSeeOther)
	default:
		return false
	}
	return true
}

func retargetLayoutCrossing(w http.ResponseWriter, r *http.Request, app bool) {
	own := "main#public-main"
	if app {
		own = "main#app-main"
	}
	if r.Header.Get("HX-Boosted") == "true" && r.Header.Get("HX-Target") != own {
		w.Header().Set("HX-Retarget", "body")
		w.Header().Set("HX-Reselect", "body")
		w.Header().Set("HX-Reswap", "outerSync")
	}
}

func layoutName(p page) string {
	if p.App {
		return "app-" + p.Variant
	}
	if p.Path == "/" {
		return "home-layout"
	}
	return "auth-" + p.Variant
}

func variant(w http.ResponseWriter, r *http.Request) string {
	if v := strings.ToUpper(r.URL.Query().Get("variant")); v != "" {
		http.SetCookie(w, &http.Cookie{Name: "variant", Value: v, Path: "/"})
		return v
	}
	if c, err := r.Cookie("variant"); err == nil {
		return c.Value
	}
	return "A"
}

func theme(r *http.Request) string {
	if c, err := r.Cookie("theme"); err == nil {
		return c.Value
	}
	return "system"
}

const templates = `
{{define "head"}}<!doctype html>
<html lang="en" class="{{if eq .Theme "dark"}}dark{{end}} h-full" data-theme="{{.Theme}}">
<head>
<meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}} · Go SSR Template</title>
<script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>
<style type="text/tailwindcss">@custom-variant dark (&:where(.dark, .dark *)); [x-cloak]{display:none}</style>
<script src="https://cdn.jsdelivr.net/npm/htmx.org@4.0.0/dist/htmx.min.js"></script>
<script>
  // stands in for app.ts: system theme, the nav component and the toast store
  if (document.documentElement.dataset.theme === 'system' && matchMedia('(prefers-color-scheme: dark)').matches) document.documentElement.classList.add('dark');
  document.addEventListener('alpine:init', () => {
    Alpine.data('nav', () => ({
      path: location.pathname,
      init() { document.addEventListener('htmx:after:swap', () => { this.path = location.pathname }) },
      active(prefix) { return this.path === prefix || this.path.startsWith(prefix + '/') },
    }));
    Alpine.store('toasts', { items: [], push(text) { const id = Date.now(); this.items.push({ id, text }); setTimeout(() => this.items = this.items.filter(t => t.id !== id), 5000) } });
  });
</script>
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.17.4/dist/cdn.min.js"></script>
</head>{{end}}

{{define "switcher"}}
<div class="fixed bottom-4 left-1/2 -translate-x-1/2 z-[100] flex items-center gap-3 rounded-full bg-fuchsia-600 text-white px-4 py-2 shadow-xl text-sm font-medium"
     x-data="{ keys: ['A','B','C'], go(d){ const i=(this.keys.indexOf('{{.Variant}}')+d+3)%3; location.search='?variant='+this.keys[i] } }"
     @keydown.window="if(['INPUT','TEXTAREA'].includes($event.target.tagName)) return; if($event.key==='ArrowLeft') go(-1); if($event.key==='ArrowRight') go(1)">
  <button @click="go(-1)" class="px-1">←</button>
  <span>PROTOTYPE {{.Variant}} · {{.VariantName}}</span>
  <button @click="go(1)" class="px-1">→</button>
</div>{{end}}

{{define "banner"}}
<div x-data="{ show: true }" x-show="show" id="realtime-banner" class="bg-amber-100 text-amber-900 dark:bg-amber-900/40 dark:text-amber-100 text-sm px-4 py-2 flex items-center justify-center gap-3">
  <span><strong>Announcement</strong> · Scheduled maintenance tonight at 22:00 UTC. <em class="opacity-60">(public channel slot)</em></span>
  <button @click="show = false" class="opacity-60 hover:opacity-100" aria-label="Dismiss">✕</button>
</div>{{end}}

{{define "toasts"}}
<div x-data class="fixed bottom-16 right-4 z-50 flex flex-col gap-2 w-80" aria-live="polite">
  <template x-for="t in $store.toasts.items" :key="t.id">
    <div class="rounded-lg border border-zinc-200 bg-white dark:bg-zinc-800 dark:border-zinc-700 shadow-lg p-4 text-sm flex gap-3">
      <span class="text-emerald-600">●</span><span x-text="t.text"></span>
    </div>
  </template>
</div>{{end}}

{{define "logo"}}<a href="/dashboard" class="flex items-center gap-2 font-semibold"><span class="grid size-8 place-items-center rounded-md bg-zinc-900 text-white dark:bg-white dark:text-zinc-900">G</span><span>Go SSR Template</span></a>{{end}}

{{define "user-menu"}}
<div x-data="{ open: false }" class="relative" @click.outside="open = false">
  <button @click="open = !open" class="flex w-full items-center gap-2 rounded-lg p-2 hover:bg-zinc-200/60 dark:hover:bg-zinc-800 text-sm">
    <span class="grid size-8 place-items-center rounded-md bg-zinc-200 dark:bg-zinc-700 font-semibold">{{.Initials}}</span>
    <span id="user-menu-name" class="truncate {{if eq .Where "header"}}hidden{{end}}">{{.Name}}</span><span class="ml-auto opacity-50">⌄</span>
  </button>
  <div x-show="open" x-transition class="absolute z-40 w-60 rounded-lg border border-zinc-200 bg-white dark:bg-zinc-900 dark:border-zinc-700 shadow-lg p-1 text-sm {{if eq .Where "bottom"}}bottom-full mb-2 left-0{{else if eq .Where "header"}}right-0 mt-2{{else}}left-0 mt-2{{end}}">
    <div class="px-2 py-2 flex gap-2 items-center"><span class="grid size-8 place-items-center rounded-md bg-zinc-200 dark:bg-zinc-700 font-semibold">{{.Initials}}</span><span><span class="block font-medium">{{.Name}}</span><span class="block text-xs text-zinc-500">{{.Email}}</span></span></div>
    <hr class="my-1 border-zinc-200 dark:border-zinc-700">
    <a href="/settings/profile" class="block rounded px-2 py-1.5 hover:bg-zinc-100 dark:hover:bg-zinc-800">⚙ Settings</a>
    <hr class="my-1 border-zinc-200 dark:border-zinc-700">
    <a href="/" class="block rounded px-2 py-1.5 hover:bg-zinc-100 dark:hover:bg-zinc-800">↪ Log out</a>
  </div>
</div>{{end}}

{{define "nav-link"}}<a href="{{.Href}}" x-bind:class="active('{{.Href}}') ? 'bg-zinc-200 dark:bg-zinc-800 font-medium' : 'hover:bg-zinc-200/60 dark:hover:bg-zinc-800/60'" class="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm">{{.Icon}} {{.Label}}</a>{{end}}

{{/* ───────────── app layouts ───────────── */}}

{{define "app-A"}}{{template "head" .}}
<body class="h-full bg-white text-zinc-900 dark:bg-zinc-900 dark:text-zinc-100" hx-boost:inherited="swap:outerSync select:#app-main target:#app-main">
<div class="flex h-full" x-data="{ mobile: false }">
  <aside x-data="nav" :class="mobile ? 'flex' : 'hidden'" class="lg:flex w-64 shrink-0 flex-col gap-4 border-r border-zinc-200 bg-zinc-50 p-4 dark:border-zinc-700 dark:bg-zinc-950 fixed inset-y-0 lg:static z-30 h-full">
    {{template "logo"}}
    <nav class="flex flex-col gap-0.5"><p class="px-2 text-xs text-zinc-500 mb-1">Platform</p>{{template "nav-link" (dict "Href" "/dashboard" "Icon" "▦" "Label" "Dashboard")}}</nav>
    <div class="mt-auto flex flex-col gap-0.5 text-sm"><a href="#" class="px-2 py-1.5">⧉ Repository</a><a href="#" class="px-2 py-1.5">☰ Documentation</a></div>
    {{template "user-menu" (dict "Where" "bottom" "Name" .UserName "Email" .UserEmail "Initials" .Initials)}}
  </aside>
  <div class="flex-1 min-w-0 overflow-y-auto">
    {{template "banner" .}}
    <header class="lg:hidden flex items-center justify-between border-b border-zinc-200 dark:border-zinc-700 p-3"><button @click="mobile = !mobile">☰</button>{{template "user-menu" (dict "Where" "header" "Name" .UserName "Email" .UserEmail "Initials" .Initials)}}</header>
    <main id="app-main" class="p-6 lg:p-8">{{if .Settings}}{{template "settings-frame-A" .}}{{else}}{{template "body" .}}{{end}}</main>
  </div>
</div>
{{template "toasts" .}}{{template "switcher" .}}
</body></html>{{end}}

{{define "app-B"}}{{template "head" .}}
<body class="h-full bg-zinc-50 text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100" hx-boost:inherited="swap:outerSync select:#app-main target:#app-main">
{{template "banner" .}}
<header x-data="nav" class="border-b border-zinc-200 bg-white dark:border-zinc-700 dark:bg-zinc-900">
  <div class="mx-auto max-w-7xl flex items-center gap-6 px-4 h-14" x-data="{ mobile: false }">
    {{template "logo"}}
    <nav class="hidden md:flex gap-1">{{template "nav-link" (dict "Href" "/dashboard" "Icon" "" "Label" "Dashboard")}}</nav>
    <div class="ml-auto flex items-center gap-3 text-sm"><a href="#" class="hidden md:inline opacity-70">Repository</a><a href="#" class="hidden md:inline opacity-70">Docs</a>
      {{template "user-menu" (dict "Where" "header" "Name" .UserName "Email" .UserEmail "Initials" .Initials)}}
    </div>
  </div>
</header>
<main id="app-main" class="mx-auto max-w-7xl p-6">{{if .Settings}}{{template "settings-frame-B" .}}{{else}}{{template "body" .}}{{end}}</main>
{{template "toasts" .}}{{template "switcher" .}}
</body></html>{{end}}

{{define "app-C"}}{{template "head" .}}
<body class="h-full bg-white text-zinc-900 dark:bg-zinc-900 dark:text-zinc-100" hx-boost:inherited="swap:outerSync select:#app-main target:#app-main">
<div class="flex h-full">
  <aside x-data="nav" class="hidden lg:flex w-64 shrink-0 flex-col gap-4 border-r border-zinc-200 bg-zinc-50 p-4 dark:border-zinc-700 dark:bg-zinc-950">
    {{template "user-menu" (dict "Where" "top" "Name" .UserName "Email" .UserEmail "Initials" .Initials)}}
    <nav class="flex flex-col gap-0.5">
      {{template "nav-link" (dict "Href" "/dashboard" "Icon" "▦" "Label" "Dashboard")}}
      <div x-data="{ open: active('/settings') }">
        <button @click="open = !open" class="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-zinc-200/60 dark:hover:bg-zinc-800/60">⚙ Settings <span class="ml-auto opacity-50" x-text="open ? '−' : '+'"></span></button>
        <div x-show="open" class="ml-4 border-l border-zinc-200 dark:border-zinc-700 pl-2 flex flex-col gap-0.5">
          {{template "nav-link" (dict "Href" "/settings/profile" "Icon" "" "Label" "Profile")}}
          {{template "nav-link" (dict "Href" "/settings/password" "Icon" "" "Label" "Password")}}
          {{template "nav-link" (dict "Href" "/settings/two-factor" "Icon" "" "Label" "Two-factor authentication")}}
          {{template "nav-link" (dict "Href" "/settings/appearance" "Icon" "" "Label" "Appearance")}}
        </div>
      </div>
    </nav>
    <div class="mt-auto">{{template "logo"}}</div>
  </aside>
  <div class="flex-1 min-w-0 overflow-y-auto">{{template "banner" .}}
  <main id="app-main" class="p-6 lg:p-10">{{if .Settings}}{{template "settings-frame-C" .}}{{else}}{{template "body" .}}{{end}}</main></div>
</div>
{{template "toasts" .}}{{template "switcher" .}}
</body></html>{{end}}

{{/* ───────────── settings frames ───────────── */}}

{{define "settings-heading"}}<h1 class="text-2xl font-semibold">Settings</h1><p class="text-zinc-500 mt-1">Manage your profile and account settings</p>{{end}}

{{define "settings-frame-A"}}
{{template "settings-heading"}}<hr class="my-6 border-zinc-200 dark:border-zinc-700">
<div class="flex flex-col md:flex-row gap-8" x-data="nav">
  <nav class="md:w-56 flex flex-col gap-0.5">
    {{template "nav-link" (dict "Href" "/settings/profile" "Icon" "" "Label" "Profile")}}
    {{template "nav-link" (dict "Href" "/settings/password" "Icon" "" "Label" "Password")}}
    {{template "nav-link" (dict "Href" "/settings/two-factor" "Icon" "" "Label" "Two-factor authentication")}}
    {{template "nav-link" (dict "Href" "/settings/appearance" "Icon" "" "Label" "Appearance")}}
  </nav>
  <div class="flex-1 max-w-lg">{{template "body" .}}</div>
</div>{{end}}

{{define "settings-frame-B"}}
{{template "settings-heading"}}
<nav class="mt-6 flex gap-6 border-b border-zinc-200 dark:border-zinc-700 text-sm" x-data="nav">
  <a href="/settings/profile" :class="active('/settings/profile') ? 'border-zinc-900 dark:border-white font-medium' : 'border-transparent opacity-70'" class="-mb-px border-b-2 pb-3">Profile</a>
  <a href="/settings/password" :class="active('/settings/password') ? 'border-zinc-900 dark:border-white font-medium' : 'border-transparent opacity-70'" class="-mb-px border-b-2 pb-3">Password</a>
  <a href="/settings/two-factor" :class="active('/settings/two-factor') ? 'border-zinc-900 dark:border-white font-medium' : 'border-transparent opacity-70'" class="-mb-px border-b-2 pb-3">Two-factor authentication</a>
  <a href="/settings/appearance" :class="active('/settings/appearance') ? 'border-zinc-900 dark:border-white font-medium' : 'border-transparent opacity-70'" class="-mb-px border-b-2 pb-3">Appearance</a>
</nav>
<div class="mt-8 max-w-xl rounded-xl border border-zinc-200 bg-white p-6 dark:border-zinc-700 dark:bg-zinc-900">{{template "body" .}}</div>{{end}}

{{define "settings-frame-C"}}
<div class="max-w-4xl"><div class="grid md:grid-cols-3 gap-8">
  <div><h1 class="text-lg font-semibold">{{.Title}}</h1><p class="text-sm text-zinc-500 mt-1">{{template "settings-blurb" .}}</p></div>
  <div class="md:col-span-2">{{template "body" .}}</div>
</div></div>{{end}}

{{define "settings-blurb"}}{{if eq .Path "/settings/profile"}}Update your name and email address{{else if eq .Path "/settings/password"}}Ensure your account is using a long, random password to stay secure{{else if eq .Path "/settings/two-factor"}}Add a second sign-in step using an authenticator app{{else}}Update the appearance settings for your account{{end}}{{end}}

{{/* ───────────── public layouts ───────────── */}}

{{define "home-layout"}}{{template "head" .}}
<body class="h-full bg-zinc-50 text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100" hx-boost:inherited="swap:outerSync select:#public-main target:#public-main">
{{template "banner" .}}
<header class="mx-auto max-w-5xl flex justify-end gap-2 p-6 text-sm"><a href="/login" class="px-4 py-1.5">Log in</a><a href="/register" class="rounded-md border border-zinc-300 dark:border-zinc-700 px-4 py-1.5">Register</a></header>
<main id="public-main" class="mx-auto max-w-5xl px-6">{{template "body" .}}</main>
{{template "switcher" .}}
</body></html>{{end}}

{{define "auth-heading"}}<div class="text-center"><h1 class="text-xl font-semibold">{{.Title}}</h1><p class="text-sm text-zinc-500 mt-1">{{template "auth-blurb" .}}</p></div>{{end}}

{{define "auth-A"}}{{template "head" .}}
<body class="h-full bg-white text-zinc-900 dark:bg-zinc-900 dark:text-zinc-100" hx-boost:inherited="swap:outerSync select:#public-main target:#public-main">
{{template "banner" .}}
<main id="public-main" class="min-h-[80vh] flex flex-col items-center justify-center p-6"><div class="w-full max-w-sm flex flex-col gap-6">
  <div class="flex justify-center">{{template "logo"}}</div>{{template "auth-heading" .}}{{template "body" .}}
</div></main>
{{template "switcher" .}}
</body></html>{{end}}

{{define "auth-B"}}{{template "head" .}}
<body class="h-full bg-zinc-100 text-zinc-900 dark:bg-zinc-950 dark:text-zinc-100" hx-boost:inherited="swap:outerSync select:#public-main target:#public-main">
{{template "banner" .}}
<main id="public-main" class="min-h-[80vh] flex flex-col items-center justify-center p-6 gap-6">
  {{template "logo"}}
  <div class="w-full max-w-md rounded-xl border border-zinc-200 bg-white p-8 shadow-sm dark:border-zinc-700 dark:bg-zinc-900 flex flex-col gap-6">{{template "auth-heading" .}}{{template "body" .}}</div>
</main>
{{template "switcher" .}}
</body></html>{{end}}

{{define "auth-C"}}{{template "head" .}}
<body class="h-full bg-white text-zinc-900 dark:bg-zinc-900 dark:text-zinc-100" hx-boost:inherited="swap:outerSync select:#public-main target:#public-main">
{{template "banner" .}}
<main id="public-main" class="min-h-screen grid lg:grid-cols-2">
  <div class="hidden lg:flex flex-col justify-between bg-zinc-900 text-white p-10"><span class="font-semibold">G · Go SSR Template</span>
    <blockquote class="text-lg">“Simplicity is prerequisite for reliability.”<footer class="text-sm opacity-60 mt-2">Edsger W. Dijkstra</footer></blockquote></div>
  <div class="flex items-center justify-center p-6"><div class="w-full max-w-sm flex flex-col gap-6">{{template "auth-heading" .}}{{template "body" .}}</div></div>
</main>
{{template "switcher" .}}
</body></html>{{end}}

{{define "auth-blurb"}}{{if eq .Path "/login"}}Enter your email and password below to log in{{else if eq .Path "/register"}}Enter your details below to create your account{{else if eq .Path "/forgot-password"}}Enter your email to receive a password reset link{{else if eq .Path "/reset-password"}}Please enter your new password below{{else if eq .Path "/verify-email"}}Please verify your email address by clicking on the link we just emailed to you{{else if eq .Path "/confirm-password"}}This is a secure area of the application. Please confirm your password before continuing{{else}}Enter the authentication code provided by your authenticator application{{end}}{{end}}

{{/* ───────────── components ───────────── */}}

{{define "field"}}<label class="flex flex-col gap-1.5 text-sm"><span class="font-medium">{{.Label}}</span>
<input type="{{.Type}}" name="{{.Name}}" value="{{.Value}}" placeholder="{{.Placeholder}}" class="rounded-md border border-zinc-300 bg-white px-3 py-2 dark:border-zinc-600 dark:bg-zinc-800 {{if .Error}}border-red-500{{end}}">
{{if .Error}}<span class="text-red-600 text-xs">{{.Error}}</span>{{end}}</label>{{end}}

{{define "button"}}<button type="{{or .Type "submit"}}" class="rounded-md px-4 py-2 text-sm font-medium {{if eq .Kind "danger"}}bg-red-600 text-white{{else if eq .Kind "ghost"}}border border-zinc-300 dark:border-zinc-600{{else}}bg-zinc-900 text-white dark:bg-white dark:text-zinc-900{{end}} {{if .Full}}w-full{{end}}">{{.Label}}</button>{{end}}

{{define "status"}}{{if .}}<p class="rounded-md bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300 text-sm text-center p-2">{{.}}</p>{{end}}{{end}}

{{define "action-message"}}{{if eq . "saved"}}<span class="text-sm text-zinc-500" x-data="{ show: true }" x-init="setTimeout(() => show = false, 2000)" x-show="show" x-transition>Saved.</span>{{end}}{{end}}

{{define "section-heading"}}<h2 class="font-semibold">{{.Title}}</h2><p class="text-sm text-zinc-500">{{.Sub}}</p>{{end}}

{{/* ───────────── page bodies (shared by all variants) ───────────── */}}

{{define "body"}}{{if eq .Body "home"}}{{template "home" .}}{{else if eq .Body "login"}}{{template "login" .}}{{else if eq .Body "register"}}{{template "register" .}}{{else if eq .Body "forgot-password"}}{{template "forgot-password" .}}{{else if eq .Body "reset-password"}}{{template "reset-password" .}}{{else if eq .Body "verify-email"}}{{template "verify-email" .}}{{else if eq .Body "confirm-password"}}{{template "confirm-password" .}}{{else if eq .Body "two-factor-challenge"}}{{template "two-factor-challenge" .}}{{else if eq .Body "dashboard"}}{{template "dashboard" .}}{{else if eq .Body "settings-profile"}}{{template "settings-profile" .}}{{else if eq .Body "settings-password"}}{{template "settings-password" .}}{{else if eq .Body "settings-two-factor"}}{{template "settings-two-factor" .}}{{else}}{{template "settings-appearance" .}}{{end}}{{end}}

{{define "home"}}<div class="grid lg:grid-cols-2 gap-8 rounded-xl border border-zinc-200 bg-white p-10 dark:border-zinc-700 dark:bg-zinc-900 mt-10">
  <div><h1 class="text-2xl font-semibold">Let's get started</h1><p class="text-zinc-500 mt-2">A server-rendered Go starter with authentication, settings and a job queue.</p>
  <ul class="mt-6 text-sm flex flex-col gap-2"><li>→ <a href="#" class="underline">Read the documentation</a></li><li>→ <a href="#" class="underline">Browse the ADRs</a></li></ul>
  <a href="/dashboard" class="inline-block mt-6 rounded-md bg-zinc-900 text-white dark:bg-white dark:text-zinc-900 px-4 py-2 text-sm">Go to dashboard (crossing)</a></div>
  <div class="rounded-lg bg-zinc-100 dark:bg-zinc-800 grid place-items-center text-6xl font-bold opacity-30">G</div>
</div>{{end}}

{{define "login"}}{{template "status" .Status}}
<form action="/dashboard" hx-boost="false" class="flex flex-col gap-5">
  {{template "field" (dict "Label" "Email address" "Type" "email" "Name" "email" "Placeholder" "email@example.com")}}
  <div class="relative">{{template "field" (dict "Label" "Password" "Type" "password" "Name" "password" "Placeholder" "Password")}}<a href="/forgot-password" class="absolute right-0 top-0 text-sm underline opacity-70">Forgot your password?</a></div>
  <label class="flex items-center gap-2 text-sm"><input type="checkbox" name="remember"> Remember me</label>
  {{template "button" (dict "Label" "Log in" "Full" true)}}
</form>
<p class="text-center text-sm text-zinc-500">Don't have an account? <a href="/register" class="underline">Sign up</a> · <a href="/two-factor-challenge" class="underline">(2FA step)</a></p>{{end}}

{{define "register"}}<form action="/dashboard" hx-boost="false" class="flex flex-col gap-5">
  {{template "field" (dict "Label" "Name" "Type" "text" "Name" "name" "Placeholder" "Full name")}}
  {{template "field" (dict "Label" "Email address" "Type" "email" "Name" "email" "Value" "taken@example.com" "Error" "The email has already been taken.")}}
  {{template "field" (dict "Label" "Password" "Type" "password" "Name" "password")}}
  {{template "field" (dict "Label" "Confirm password" "Type" "password" "Name" "password_confirmation")}}
  {{template "button" (dict "Label" "Create account" "Full" true)}}
</form>
<p class="text-center text-sm text-zinc-500">Already have an account? <a href="/login" class="underline">Log in</a></p>{{end}}

{{define "forgot-password"}}{{template "status" .Status}}
<form action="/forgot-password" class="flex flex-col gap-5"><input type="hidden" name="status" value="If an account exists, a reset link has been sent.">
  {{template "field" (dict "Label" "Email address" "Type" "email" "Name" "email" "Placeholder" "email@example.com")}}
  {{template "button" (dict "Label" "Email password reset link" "Full" true)}}
</form>
<p class="text-center text-sm text-zinc-500">Or, return to <a href="/login" class="underline">log in</a> · <a href="/reset-password" class="underline">(open reset link)</a></p>{{end}}

{{define "reset-password"}}<form action="/login" class="flex flex-col gap-5"><input type="hidden" name="status" value="Your password has been reset.">
  {{template "field" (dict "Label" "Email" "Type" "email" "Name" "email" "Value" "danny@example.com")}}
  {{template "field" (dict "Label" "Password" "Type" "password" "Name" "password")}}
  {{template "field" (dict "Label" "Confirm password" "Type" "password" "Name" "password_confirmation")}}
  {{template "button" (dict "Label" "Reset password" "Full" true)}}
</form>{{end}}

{{define "verify-email"}}{{template "status" .Status}}
<form action="/verify-email" class="flex flex-col gap-3 items-center"><input type="hidden" name="status" value="A new verification link has been sent to your email address.">
  {{template "button" (dict "Label" "Resend verification email" "Full" true)}}
  <a href="/" class="text-sm underline opacity-70">Log out</a>
</form>{{end}}

{{define "confirm-password"}}<form action="/settings/two-factor" class="flex flex-col gap-5">
  {{template "field" (dict "Label" "Password" "Type" "password" "Name" "password")}}
  {{template "button" (dict "Label" "Confirm" "Full" true)}}
</form>{{end}}

{{define "two-factor-challenge"}}<form action="/dashboard" hx-boost="false" class="flex flex-col gap-5" x-data="{ recovery: false }">
  <div x-show="!recovery">{{template "field" (dict "Label" "Code" "Type" "text" "Name" "code" "Placeholder" "123456")}}</div>
  <div x-show="recovery" x-cloak>{{template "field" (dict "Label" "Recovery code" "Type" "text" "Name" "recovery_code" "Placeholder" "abcd-efgh-ijkl")}}</div>
  {{template "button" (dict "Label" "Continue" "Full" true)}}
  <p class="text-center text-sm text-zinc-500">or you can <button type="button" class="underline" @click="recovery = !recovery" x-text="recovery ? 'log in using an authentication code' : 'log in using a recovery code'"></button></p>
</form>{{end}}

{{define "dashboard"}}<div class="flex flex-col gap-4">
  <div class="grid md:grid-cols-3 gap-4">{{range $i := (dict "a" 1 "b" 2 "c" 3)}}<div class="aspect-video rounded-xl border border-dashed border-zinc-300 dark:border-zinc-700 bg-[repeating-linear-gradient(45deg,transparent,transparent_8px,rgba(120,120,120,.08)_8px,rgba(120,120,120,.08)_16px)]"></div>{{end}}</div>
  <div class="h-80 rounded-xl border border-dashed border-zinc-300 dark:border-zinc-700 grid place-items-center text-sm text-zinc-500">
    <div class="flex flex-col items-center gap-3"><span>Placeholder content proving the auth + verified middleware</span>
    <button x-data @click="$store.toasts.push('Two-factor authentication was disabled from another session.')" class="rounded-md border border-zinc-300 dark:border-zinc-600 px-3 py-1.5">Simulate private-channel toast</button>
    <a href="/" class="underline">Go home (layout crossing)</a></div>
  </div>
</div>{{end}}

{{define "settings-profile"}}<div class="flex flex-col gap-10">
<form action="/settings/profile" class="flex flex-col gap-5"><input type="hidden" name="status" value="saved">
  {{if ne .Variant "C"}}{{template "section-heading" (dict "Title" "Profile" "Sub" "Update your name and email address")}}{{end}}
  {{template "field" (dict "Label" "Name" "Type" "text" "Name" "name" "Value" .UserName)}}
  {{template "field" (dict "Label" "Email" "Type" "email" "Name" "email" "Value" .UserEmail)}}
  <p class="text-sm text-zinc-500">Your email address is unverified. <a href="/verify-email?status=A+new+verification+link+has+been+sent." class="underline">Click here to re-send the verification email.</a></p>
  <div class="flex items-center gap-4">{{template "button" (dict "Label" "Save")}}{{template "action-message" .Status}}</div>
</form>
<section x-data="{ open: false }" class="flex flex-col gap-4 {{if eq .Variant "C"}}rounded-xl border border-red-200 dark:border-red-900 p-5{{end}}">
  {{template "section-heading" (dict "Title" "Delete account" "Sub" "Delete your account and all of its resources")}}
  <div><button @click="open = true" class="rounded-md bg-red-600 text-white px-4 py-2 text-sm font-medium">Delete account</button></div>
  {{template "modal-delete-account"}}
</section></div>{{end}}

{{define "modal-delete-account"}}<div x-show="open" x-cloak class="fixed inset-0 z-50 grid place-items-center bg-black/50 p-4" @keydown.escape.window="open = false">
  <form action="/" hx-boost="false" @click.outside="open = false" class="w-full max-w-md rounded-xl bg-white dark:bg-zinc-900 p-6 flex flex-col gap-5 shadow-xl">
    <div><h3 class="text-lg font-semibold">Are you sure you want to delete your account?</h3><p class="text-sm text-zinc-500 mt-1">Once your account is deleted, all of its resources and data will be permanently deleted. Please enter your password to confirm.</p></div>
    {{template "field" (dict "Label" "Password" "Type" "password" "Name" "password")}}
    <div class="flex justify-end gap-2"><button type="button" @click="open = false" class="rounded-md border border-zinc-300 dark:border-zinc-600 px-4 py-2 text-sm">Cancel</button>{{template "button" (dict "Label" "Delete account" "Kind" "danger")}}</div>
  </form>
</div>{{end}}

{{define "settings-password"}}<form action="/settings/password" class="flex flex-col gap-5"><input type="hidden" name="status" value="saved">
  {{if ne .Variant "C"}}{{template "section-heading" (dict "Title" "Update password" "Sub" "Ensure your account is using a long, random password to stay secure")}}{{end}}
  {{template "field" (dict "Label" "Current password" "Type" "password" "Name" "current_password" "Error" "The password is incorrect.")}}
  {{template "field" (dict "Label" "New password" "Type" "password" "Name" "password")}}
  {{template "field" (dict "Label" "Confirm password" "Type" "password" "Name" "password_confirmation")}}
  <div class="flex items-center gap-4">{{template "button" (dict "Label" "Save")}}{{template "action-message" .Status}}</div>
</form>{{end}}

{{define "settings-two-factor"}}<div class="flex flex-col gap-6">
  {{if ne .Variant "C"}}{{template "section-heading" (dict "Title" "Two-factor authentication" "Sub" "Manage your two-factor authentication settings")}}{{end}}
  {{if .TwoFactor}}
    <div><span class="rounded-full bg-emerald-100 text-emerald-700 dark:bg-emerald-900/40 dark:text-emerald-300 px-2 py-0.5 text-xs font-medium">Enabled</span></div>
    <p class="text-sm text-zinc-500">With two-factor authentication enabled, you will be prompted for a secure, random code during login, which you can retrieve from the authenticator app on your phone.</p>
    <div x-data="{ show: false }" class="rounded-xl border border-zinc-200 dark:border-zinc-700 p-5 flex flex-col gap-4">
      <div class="flex items-center justify-between"><div><h3 class="font-medium">Recovery codes</h3><p class="text-sm text-zinc-500">{{len .RecoveryCodes}} of 8 remaining</p></div>
        <div class="flex gap-2"><button @click="show = !show" class="rounded-md border border-zinc-300 dark:border-zinc-600 px-3 py-1.5 text-sm" x-text="show ? 'Hide recovery codes' : 'View recovery codes'"></button><a href="/settings/two-factor?status=regenerated" class="rounded-md border border-zinc-300 dark:border-zinc-600 px-3 py-1.5 text-sm">Regenerate</a></div></div>
      <div x-show="show" x-cloak class="grid grid-cols-2 gap-1 rounded-md bg-zinc-100 dark:bg-zinc-800 p-4 font-mono text-sm">{{range .RecoveryCodes}}<span>{{.}}</span>{{end}}</div>
      <p x-show="show" x-cloak class="text-xs text-zinc-500">Each recovery code can be used once. Store them in a password manager.</p>
    </div>
    <div><a href="/settings/two-factor?two-factor=off" hx-boost="false" class="rounded-md bg-red-600 text-white px-4 py-2 text-sm font-medium">Disable 2FA</a></div>
  {{else}}
    <div><span class="rounded-full bg-zinc-200 dark:bg-zinc-700 px-2 py-0.5 text-xs font-medium">Disabled</span></div>
    <p class="text-sm text-zinc-500">When you enable two-factor authentication, you will be prompted for a secure code during login. This code can be retrieved from a TOTP-supported application on your phone.</p>
    <div x-data="{ open: false }"><button @click="open = true" class="rounded-md bg-zinc-900 text-white dark:bg-white dark:text-zinc-900 px-4 py-2 text-sm font-medium">Enable 2FA</button>
      <div x-show="open" x-cloak class="fixed inset-0 z-50 grid place-items-center bg-black/50 p-4" @keydown.escape.window="open = false">
        <form action="/settings/two-factor" hx-boost="false" @click.outside="open = false" class="w-full max-w-md rounded-xl bg-white dark:bg-zinc-900 p-6 flex flex-col gap-5 shadow-xl items-center text-center"><input type="hidden" name="two-factor" value="on">
          <h3 class="text-lg font-semibold">Enable two-factor authentication</h3><p class="text-sm text-zinc-500">Scan the QR code or enter the setup key in your authenticator app, then enter the 6-digit code.</p>
          <div class="size-40 rounded-md bg-[conic-gradient(#000_25%,#fff_0_50%,#000_0_75%,#fff_0)] bg-[length:16px_16px] border-8 border-white"></div>
          <code class="text-xs rounded bg-zinc-100 dark:bg-zinc-800 px-2 py-1">JBSW Y3DP EHPK 3PXP</code>
          <div class="w-full text-left">{{template "field" (dict "Label" "Code" "Type" "text" "Name" "code" "Placeholder" "123456")}}</div>
          <div class="flex gap-2 w-full"><button type="button" @click="open = false" class="flex-1 rounded-md border border-zinc-300 dark:border-zinc-600 px-4 py-2 text-sm">Cancel</button><button class="flex-1 rounded-md bg-zinc-900 text-white dark:bg-white dark:text-zinc-900 px-4 py-2 text-sm">Confirm</button></div>
        </form>
      </div>
    </div>
  {{end}}
  <p class="text-xs text-zinc-500">Prototype: enable/disable sit behind <a href="/confirm-password" class="underline">password confirmation</a> (layout crossing).</p>
</div>{{end}}

{{define "settings-appearance"}}<div class="flex flex-col gap-5">
  {{if ne .Variant "C"}}{{template "section-heading" (dict "Title" "Appearance" "Sub" "Update the appearance settings for your account")}}{{end}}
  <div class="inline-flex rounded-lg bg-zinc-100 dark:bg-zinc-800 p-1 text-sm w-fit">
    <a href="?theme=light" hx-boost="false" class="rounded-md px-4 py-1.5 {{if eq .Theme "light"}}bg-white dark:bg-zinc-700 shadow-sm{{end}}">☀ Light</a>
    <a href="?theme=dark" hx-boost="false" class="rounded-md px-4 py-1.5 {{if eq .Theme "dark"}}bg-white dark:bg-zinc-700 shadow-sm{{end}}">☾ Dark</a>
    <a href="?theme=system" hx-boost="false" class="rounded-md px-4 py-1.5 {{if eq .Theme "system"}}bg-white dark:bg-zinc-700 shadow-sm{{end}}">▣ System</a>
  </div>
</div>{{end}}
`
