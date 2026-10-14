package asset

import (
	"io/fs"
	"net/http"
	"strings"
)

type manifestSource interface {
	load() (manifest, error)
}

type Server struct {
	files     fs.FS
	manifests manifestSource
	caching   cachePolicy
}

func NewImmutable(files fs.FS) (*Server, error) {
	loaded, err := readManifest(files)
	if err != nil {
		return nil, err
	}

	return &Server{files: files, manifests: fixedManifest{manifest: loaded}, caching: immutable()}, nil
}

func NewLive(files fs.FS) *Server {
	return &Server{files: files, manifests: liveManifest{files: files}, caching: uncached()}
}

func (s *Server) ScriptURL(source string) (string, error) {
	found, err := s.entry(source)
	if err != nil {
		return "", err
	}

	return url(found.File), nil
}

func (s *Server) StylesheetURLs(source string) ([]string, error) {
	found, err := s.entry(source)
	if err != nil {
		return nil, err
	}

	urls := make([]string, 0, len(found.CSS))
	for _, file := range found.CSS {
		urls = append(urls, url(file))
	}

	return urls, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	loaded, err := s.manifests.load()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	file, found := loaded.resolve(strings.TrimPrefix(r.URL.Path, "/"))
	if !found {
		http.NotFound(w, r)

		return
	}

	w.Header().Set("Cache-Control", s.caching.cacheControl(file))
	http.ServeFileFS(w, r, s.files, file.file)
}

func (s *Server) entry(source string) (chunk, error) {
	loaded, err := s.manifests.load()
	if err != nil {
		return chunk{}, err
	}

	return loaded.entry(source)
}

func url(file string) string {
	return "/" + file
}
