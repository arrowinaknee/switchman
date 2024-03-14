package http

import (
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// Server is a host container for endpoints
type Server struct {
	Endpoints []Endpoint
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path
	for _, e := range s.Endpoints {
		if strings.HasPrefix(path, e.Location) {
			e.handle(w, r)
			return
		}
	}
	respondWith404(w, r)
}

// An endpoint is the main unit of routing inside the server. Incoming requests
// are handled by corresponding EndpointFunction
type Endpoint struct {
	Location string
	Function EndpointFunction
}

func (ep *Endpoint) handle(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path
	// TODO: customize trimming
	var localPath = strings.TrimPrefix(path, ep.Location)

	ep.Function.Serve(w, r, localPath)
}

// An endpoint function provides the action that will be applied to requests
// received by its parent endpoint
type EndpointFunction interface {
	Serve(w http.ResponseWriter, r *http.Request, localPath string)
}

// EndpointFiles is an endpoint function that serves files from local filesystem
type EndpointFiles struct {
	FileRoot string // Path to the directory that the files will be served from
}

func (f *EndpointFiles) Serve(w http.ResponseWriter, r *http.Request, localPath string) {
	// only serve files from current subtree to prevent access to the whole filesystem
	if !filepath.IsLocal(localPath) && localPath != "" {
		respondWith404(w, r)
		return
	}
	// TODO: use a setting
	if localPath == "" {
		localPath = "index.html"
	}

	// join will remove all excess separators
	var fpath = filepath.Join(f.FileRoot, localPath)
	var file, err = os.OpenFile(fpath, os.O_RDONLY, 0)
	if err != nil {
		respondWith404(w, r)
		return
	}
	defer file.Close()
	var mime = mime.TypeByExtension(filepath.Ext(fpath))
	w.Header().Set("Content-Type", mime)
	io.Copy(w, file)
}

// EndpointRedirect is an endpoint function that sends a redirect response
type EndpointRedirect struct {
	Target string
}

func (f *EndpointRedirect) Serve(w http.ResponseWriter, r *http.Request, localPath string) {
	http.Redirect(w, r, f.Target, http.StatusMovedPermanently)
}
