package asset

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed all:dist
var embedded embed.FS

func Embedded() (fs.FS, error) {
	dist, err := fs.Sub(embedded, "dist")
	if err != nil {
		return nil, fmt.Errorf("open the embedded assets: %w", err)
	}

	return dist, nil
}
