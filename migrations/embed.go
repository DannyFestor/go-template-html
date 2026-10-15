package migrations

import "embed"

//go:generate go run ../internal/migrationsum/cmd/write

//go:embed *.sql migrations.sum
var FS embed.FS
