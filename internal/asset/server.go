package asset

import (
	"io/fs"
	"net/http"
	"strings"
)

type manifestSource interface {
	load() (manifest, error)
}

type fileServerSource interface {
	load() http.Handler
}

type Server struct {
	manifests   manifestSource
	fileServers fileServerSource
	caching     cachePolicy
}

func NewImmutable(files fs.FS) (*Server, error) {
	loaded, err := readManifest(files)
	if err != nil {
		return nil, err
	}

	return &Server{
		manifests:   fixedManifest{manifest: loaded},
		fileServers: fixedFileServer{handler: precompressedFileServer(files)},
		caching:     fixedBuildCaching(),
	}, nil
}

func NewLive(files fs.FS) *Server {
	return &Server{
		manifests:   liveManifest{files: files},
		fileServers: liveFileServer{files: files},
		caching:     liveBuildCaching(),
	}
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
	s.fileServers.load().ServeHTTP(w, requestFor(r, file.file))
}

func (s *Server) entry(source string) (chunk, error) {
	loaded, err := s.manifests.load()
	if err != nil {
		return chunk{}, err
	}

	return loaded.entry(source)
}

func requestFor(r *http.Request, file string) *http.Request {
	resolved := r.Clone(r.Context())
	resolved.URL.Path = url(file)
	resolved.URL.RawPath = ""

	return resolved
}

func url(file string) string {
	return "/" + file
}
