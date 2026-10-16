package asset

import (
	"fmt"
	"io/fs"
	"net/http"

	"github.com/vearutop/statigz"
	"github.com/vearutop/statigz/brotli"
)

type fixedFileServer struct {
	handler http.Handler
}

func (f fixedFileServer) load() http.Handler {
	return f.handler
}

// liveFileServer indexes the files on every request, because each rebuild of `make watch-vite` renames them.
type liveFileServer struct {
	files fs.FS
}

func (l liveFileServer) load() http.Handler {
	return precompressedFileServer(l.files)
}

func precompressedFileServer(files fs.FS) http.Handler {
	return statigz.FileServer(readDirFS{FS: files}, brotli.AddEncoding)
}

// readDirFS exists because statigz takes an fs.ReadDirFS, and fs.Sub and os.DirFS only promise an fs.FS.
type readDirFS struct {
	fs.FS
}

func (r readDirFS) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := fs.ReadDir(r.FS, name)
	if err != nil {
		return nil, fmt.Errorf("list the assets in %s: %w", name, err)
	}

	return entries, nil
}
