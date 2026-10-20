module github.com/DannyFestor/go-template-html

go 1.27

toolchain go1.27.2

require (
	github.com/PuerkitoBio/goquery v1.13.0
	github.com/a-h/templ v0.3.1070
	github.com/brianvoe/gofakeit/v7 v7.18.0
	github.com/caarlos0/env/v11 v11.4.1
	github.com/jackc/pgx/v5 v5.11.0
	github.com/joho/godotenv v1.5.1
	github.com/klauspost/compress v1.20.3
	github.com/mxschmitt/playwright-go v0.6201.1
	github.com/pressly/goose/v3 v3.29.0
	github.com/redis/go-redis/v9 v9.23.0
	github.com/stretchr/testify v1.12.1
	github.com/vearutop/statigz v1.5.0
	go.uber.org/mock v0.6.0
	golang.org/x/crypto v0.58.0
)

require (
	github.com/a-h/parse v0.0.0-20250122154542-74294addb73e // indirect
	github.com/andybalholm/brotli v1.2.6 // indirect
	github.com/andybalholm/cascadia v1.3.4 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cli/browser v1.3.0 // indirect
	github.com/deckarep/golang-set/v2 v2.8.0 // indirect
	github.com/fatih/color v1.19.0 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/go-stack/stack v1.8.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.3 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mfridman/interpolate v0.0.2 // indirect
	github.com/natefinch/atomic v1.0.1 // indirect
	github.com/sethvargo/go-retry v0.5.0 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/net v0.61.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	golang.org/x/tools v0.51.0 // indirect
)

tool (
	github.com/a-h/templ/cmd/templ
	github.com/mxschmitt/playwright-go/cmd/playwright
	go.uber.org/mock/mockgen
)
