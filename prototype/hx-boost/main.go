// PROTOTYPE, throwaway. Answers one question for the "Interactivity layer" ticket:
// does app-wide hx-boost feel worth it compared with targeted htmx + full page loads?
//
// Three variants, switchable from the floating bar (or ←/→), persisted in a cookie:
//
//	A  full page loads for navigation, htmx only on the profile form
//	B  hx-boost on <body>: links/forms swap the whole body (wire:navigate equivalent)
//	C  hx-boost swapping only <main>: sidebar DOM (and its Alpine state) survives navigation
//
// /home and /about use a second (public) layout to test boosted layout crossings,
// /expired a redirect crossing.
//
// Verdict: C, plus the server-side retarget on layout crossings (see ADR-0004).
//
// Run: go run ./prototype/hx-boost/main.go  →  http://localhost:8099
package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

var variants = []struct{ Key, Name string }{
	{"A", "Full page loads + targeted htmx"},
	{"B", "hx-boost, whole body"},
	{"C", "hx-boost, <main> only"},
	{"D", "Full page loads + persisted sidebar state"},
}

type page struct {
	Variant     string
	VariantName string
	DelayMS     int
	Path        string
	Title       string
	RenderedAt  string
	Name        string
	Saved       bool
	Error       string
	User        string
	Layout      string
}

var profileName = "Danny"

func main() {
	tpl := template.Must(template.New("").Funcs(template.FuncMap{
		"delays": func() []int { return []int{0, 150, 500, 1200} },
		"themes": func() []string { return []string{"light", "dark", "system"} },
		"seq": func() []int {
			rows := make([]int, 40)
			for i := range rows {
				rows[i] = i + 1
			}
			return rows
		},
	}).Parse(templates))
	mux := http.NewServeMux()

	render := func(w http.ResponseWriter, r *http.Request, name string, p page) {
		p.Variant, p.VariantName = currentVariant(r)
		p.DelayMS = currentDelay(r)
		p.Path = r.URL.Path
		p.RenderedAt = time.Now().Format("15:04:05.000")
		p.User = profileName
		retargetOnLayoutCrossing(w, r, p.Layout)
		if err := tpl.ExecuteTemplate(w, name, p); err != nil {
			log.Println(err)
		}
	}

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	})
	mux.HandleFunc("GET /dashboard", func(w http.ResponseWriter, r *http.Request) {
		render(w, r, "dashboard", page{Title: "Dashboard"})
	})
	mux.HandleFunc("GET /settings/profile", func(w http.ResponseWriter, r *http.Request) {
		render(w, r, "profile", page{Title: "Profile", Name: profileName, Saved: r.URL.Query().Has("saved")})
	})
	mux.HandleFunc("POST /settings/profile", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		p := page{Title: "Profile", Name: name}
		if name == "" {
			p.Error = "The name field is required."
		} else {
			profileName, p.Saved = name, true
		}
		if r.Header.Get("HX-Request") == "true" && r.FormValue("_fragment") == "1" {
			render(w, r, "profile-form", p)
			return
		}
		if p.Error != "" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			render(w, r, "profile", p)
			return
		}
		http.Redirect(w, r, "/settings/profile?saved=1", http.StatusSeeOther)
	})
	mux.HandleFunc("GET /settings/appearance", func(w http.ResponseWriter, r *http.Request) {
		render(w, r, "appearance", page{Title: "Appearance"})
	})
	mux.HandleFunc("GET /home", func(w http.ResponseWriter, r *http.Request) {
		render(w, r, "public", page{Title: "Home", Layout: "public"})
	})
	mux.HandleFunc("GET /about", func(w http.ResponseWriter, r *http.Request) {
		render(w, r, "public", page{Title: "About", Layout: "public"})
	})
	// Stands in for any middleware redirect that leaves the authenticated layout.
	mux.HandleFunc("GET /expired", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/home", http.StatusFound)
	})
	mux.HandleFunc("GET /long", func(w http.ResponseWriter, r *http.Request) {
		render(w, r, "long", page{Title: "Long page"})
	})

	log.Println("prototype on http://localhost:8099")
	log.Fatal(http.ListenAndServe("localhost:8099", logRequests(simulateLatency(mux))))
}

// A boosted request aimed at the other layout's <main> gets the whole body swapped instead.
func retargetOnLayoutCrossing(w http.ResponseWriter, r *http.Request, layout string) {
	ownTarget := "main#main"
	if layout == "public" {
		ownTarget = "main#public-main"
	}
	if r.Header.Get("HX-Boosted") == "true" && r.Header.Get("HX-Target") != ownTarget {
		w.Header().Set("HX-Retarget", "body")
		w.Header().Set("HX-Reselect", "body")
		w.Header().Set("HX-Reswap", "outerSync")
	}
}

func currentVariant(r *http.Request) (string, string) {
	key := r.URL.Query().Get("variant")
	if key == "" {
		if c, err := r.Cookie("variant"); err == nil {
			key = c.Value
		}
	}
	for _, v := range variants {
		if v.Key == key {
			return v.Key, v.Name
		}
	}
	return variants[0].Key, variants[0].Name
}

func currentDelay(r *http.Request) int {
	if c, err := r.Cookie("delay"); err == nil {
		if ms, err := strconv.Atoi(c.Value); err == nil {
			return ms
		}
	}
	return 150
}

// Localhost is unrealistically fast; the delay stands in for network + DB time.
func simulateLatency(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Duration(currentDelay(r)) * time.Millisecond)
		next.ServeHTTP(w, r)
	})
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		kind := "full"
		if r.Header.Get("HX-Request") == "true" {
			kind = "htmx"
		}
		log.Printf("%-4s %-5s %s boosted=%q type=%q target=%q source=%q", kind, r.Method, r.URL.Path,
			r.Header.Get("HX-Boosted"), r.Header.Get("HX-Request-Type"), r.Header.Get("HX-Target"), r.Header.Get("HX-Source"))
		next.ServeHTTP(w, r)
	})
}

const templates = `
{{define "head"}}<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}} · prototype</title>
<script src="https://cdn.jsdelivr.net/npm/htmx.org@4.0.0/dist/htmx.min.js"></script>
<script defer src="https://cdn.jsdelivr.net/npm/@alpinejs/persist@3.17.4/dist/cdn.min.js"></script>
<script defer src="https://cdn.jsdelivr.net/npm/alpinejs@3.17.4/dist/cdn.min.js"></script>
<style>
  * { box-sizing: border-box }
  body { margin: 0; font: 15px/1.5 system-ui, sans-serif; color: #18181b; background: #fafafa; display: flex; min-height: 100vh }
  aside { width: 240px; background: #f4f4f5; border-right: 1px solid #e4e4e7; padding: 16px; flex-shrink: 0 }
  aside a { display: block; padding: 6px 10px; border-radius: 6px; color: inherit; text-decoration: none }
  aside a:hover { background: #e4e4e7 }
  aside a.active { background: #fff; font-weight: 600; box-shadow: 0 1px 2px #0001 }
  aside button { width: 100%; text-align: left; background: none; border: 0; padding: 6px 10px; font: inherit; cursor: pointer; color: #52525b }
  main { flex: 1; padding: 32px 40px 120px }
  h1 { margin-top: 0 }
  .card { background: #fff; border: 1px solid #e4e4e7; border-radius: 10px; padding: 20px; margin-bottom: 16px; max-width: 560px }
  .state { font: 12px/1.6 ui-monospace, monospace; background: #fff; border: 1px dashed #a1a1aa; border-radius: 8px; padding: 10px 14px; max-width: 560px; margin-bottom: 24px }
  input[type=text] { width: 100%; padding: 8px 10px; border: 1px solid #d4d4d8; border-radius: 6px; font: inherit }
  .btn { background: #18181b; color: #fff; border: 0; border-radius: 6px; padding: 8px 14px; font: inherit; cursor: pointer; margin-top: 10px }
  .err { color: #dc2626; font-size: 13px } .ok { color: #16a34a; font-size: 13px; margin-left: 10px }
  #progress { position: fixed; top: 0; left: 0; height: 3px; width: 100%; background: linear-gradient(90deg,#6366f1,#ec4899); display: none; animation: grow 1s ease-out }
  body:has(.htmx-request) #progress { display: block }
  @keyframes grow { from { width: 0 } to { width: 90% } }
  #switcher { position: fixed; bottom: 16px; left: 50%; transform: translateX(-50%); background: #18181b; color: #fff; border-radius: 999px; padding: 6px 10px; display: flex; gap: 10px; align-items: center; box-shadow: 0 6px 24px #0004; font-size: 13px; z-index: 50 }
  #switcher button, #switcher select { background: #3f3f46; color: #fff; border: 0; border-radius: 999px; padding: 4px 10px; cursor: pointer; font: inherit }
</style>
</head>
{{end}}

{{define "open"}}{{template "head" .}}
<body data-layout="app" {{if eq .Variant "B"}}hx-boost:inherited="true"{{end}}{{if eq .Variant "C"}}hx-boost:inherited="swap:outerSync select:#main target:#main"{{end}}>
<div id="progress"></div>
<aside {{if eq .Variant "D"}}x-data="{ settingsOpen: $persist(false).using(sessionStorage), clicks: $persist(0).using(sessionStorage) }"{{else}}x-data="{ settingsOpen: false, clicks: 0 }"{{end}}>
  <strong style="display:block;margin:4px 10px 16px">Go SSR template</strong>
  <nav id="nav-main">
  <a href="/dashboard" class="{{if eq .Path "/dashboard"}}active{{end}}">Dashboard</a>
  <a href="/long" class="{{if eq .Path "/long"}}active{{end}}">Long page</a>
  <a href="/home">→ Public home (crossing)</a>
  <a href="/expired">→ Expired session (redirect crossing)</a>
  </nav>
  <button @click="settingsOpen = !settingsOpen; clicks++">Settings <span x-text="settingsOpen ? '▾' : '▸'"></span></button>
  <div x-show="settingsOpen" style="padding-left:12px">
    <nav id="nav-settings">
    <a href="/settings/profile" class="{{if eq .Path "/settings/profile"}}active{{end}}">Profile</a>
    <a href="/settings/appearance" class="{{if eq .Path "/settings/appearance"}}active{{end}}">Appearance</a>
    </nav>
  </div>
  <p style="font-size:12px;color:#71717a;margin:16px 10px">Signed in as <b id="sidebar-user">{{.User}}</b></p>
  <p style="font-size:12px;color:#71717a;margin:16px 10px">Sidebar Alpine state: open=<span x-text="settingsOpen"></span>, toggles=<span x-text="clicks"></span><br>sidebar rendered {{.RenderedAt}}</p>
</aside>
<main id="main">
<div class="state">
  variant: <b>{{.Variant}}</b> ({{.VariantName}}) · delay {{.DelayMS}}ms<br>
  &lt;main&gt; rendered at: {{.RenderedAt}}<br>
  full page loads this tab: <b id="full-loads">?</b> · htmx swaps since last full load: <b id="swaps">0</b>
</div>
{{end}}

{{define "close"}}
</main>
{{template "switcher" .}}
{{end}}

{{define "switcher"}}
<div id="switcher">
  <button onclick="cycle(-1)">←</button>
  <span>{{.Variant}} ({{.VariantName}})</span>
  <button onclick="cycle(1)">→</button>
  <select onchange="document.cookie='delay='+this.value+';path=/';location.reload()">
    {{range $ms := delays}}<option value="{{$ms}}" {{if eq $ms $.DelayMS}}selected{{end}}>{{$ms}}ms</option>{{end}}
  </select>
</div>
<script>
  if (!window.__prototypeBooted) {
    window.__prototypeBooted = true;
    const n = +(sessionStorage.fullLoads || 0) + 1;
    sessionStorage.fullLoads = n;
    let swaps = 0;
    const paint = () => {
      const f = document.getElementById('full-loads'), s = document.getElementById('swaps');
      if (f) f.textContent = n;
      if (s) s.textContent = swaps;
    };
    paint();
    document.addEventListener('htmx:after:swap', () => { swaps++; paint(); });
    window.cycle = (dir) => {
      const keys = ['A','B','C','D'];
      const cur = '{{.Variant}}';
      const next = keys[(keys.indexOf(cur) + dir + keys.length) % keys.length];
      document.cookie = 'variant=' + next + ';path=/';
      location.reload();
    };
    document.addEventListener('keydown', (e) => {
      if (e.target.closest('input,textarea,select,[contenteditable]')) return;
      if (e.key === 'ArrowLeft') cycle(-1);
      if (e.key === 'ArrowRight') cycle(1);
    });
  }
</script>
</body></html>
{{end}}

{{define "dashboard"}}{{template "open" .}}
<h1>Dashboard</h1>
<div class="card">Click around the sidebar. Watch the address bar, the progress bar at the top, whether the page flashes white, and whether the "Settings" group stays open.</div>
<div class="card">Things to try per variant: navigate; open Settings then navigate; scroll the long page then go back; save the profile form empty and filled; use the browser back button.</div>
{{template "close" .}}{{end}}

{{define "long"}}{{template "open" .}}
<h1>Long page</h1>
<p>Scroll down, click a link at the bottom, then press Back: is your scroll position restored?</p>
{{range seq}}<div class="card">Row {{.}}</div>{{end}}
<a href="/dashboard">→ Dashboard</a>
{{template "close" .}}{{end}}

{{define "profile"}}{{template "open" .}}
<h1>Profile</h1>
{{template "profile-form" .}}
<div class="card">The form always uses targeted htmx (hx-post, swaps only the form, opts out of boost) and still works without JS (POST → redirect).</div>
{{template "close" .}}{{end}}

{{define "profile-form"}}
<form id="profile-form" class="card" method="post" action="/settings/profile"
      hx-boost="false" hx-post="/settings/profile" hx-target="this" hx-swap="outerHTML">
  <input type="hidden" name="_fragment" value="1">
  <label>Name <input type="text" name="name" value="{{.Name}}"></label>
  {{if .Error}}<div class="err">{{.Error}}</div>{{end}}
  <button class="btn">Save</button>{{if .Saved}}<span class="ok" x-data x-init="setTimeout(() => $el.remove(), 2000)">Saved.</span>{{end}}
  <div style="font-size:12px;color:#71717a;margin-top:8px">form rendered {{.RenderedAt}}</div>
</form>
{{if .Saved}}<b id="sidebar-user" hx-swap-oob="true">{{.User}}</b>{{end}}
{{end}}

{{define "public"}}{{template "head" .}}
<body data-layout="public" hx-boost:inherited="swap:outerSync select:#public-main target:#public-main" style="display:block">
<div id="progress"></div>
<header style="display:flex;gap:16px;padding:16px 40px;background:#eef2ff;border-bottom:1px solid #c7d2fe">
  <strong>Public layout</strong>
  <a href="/home">Home</a><a href="/about">About</a><a href="/dashboard">Dashboard → (crossing)</a>
</header>
<main id="public-main">
<div class="state">
  public layout · &lt;main&gt; rendered at: {{.RenderedAt}}<br>
  full page loads this tab: <b id="full-loads">?</b> · htmx swaps since last full load: <b id="swaps">0</b>
</div>
<h1>{{.Title}}</h1>
<div class="card">Header rendered once per full load; boosted links swap only #public-main.</div>
</main>
{{template "switcher" .}}
{{end}}

{{define "appearance"}}{{template "open" .}}
<h1>Appearance</h1>
<div class="card" x-data="{ theme: 'system' }">
  Alpine-only state on the page (resets on every navigation in all variants):
  <p>{{range $t := themes}}<label style="margin-right:12px"><input type="radio" value="{{$t}}" x-model="theme"> {{$t}}</label>{{end}}</p>
  <p>Selected: <b x-text="theme"></b></p>
</div>
{{template "close" .}}{{end}}
`
