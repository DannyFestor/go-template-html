package bootstrap

import (
	"fmt"
	"os"

	"github.com/DannyFestor/go-template-html/internal/asset"
)

const devAssetDir = "internal/asset/dist"

func newAssetServer(environment Environment) (*asset.Server, error) {
	if environment != EnvironmentProduction {
		return asset.NewLive(os.DirFS(devAssetDir)), nil
	}

	return newEmbeddedAssetServer()
}

func newEmbeddedAssetServer() (*asset.Server, error) {
	embedded, err := asset.Embedded()
	if err != nil {
		return nil, fmt.Errorf("serve the embedded assets: %w", err)
	}

	server, err := asset.NewImmutable(embedded)
	if err != nil {
		return nil, fmt.Errorf("serve the embedded assets: %w", err)
	}

	return server, nil
}
