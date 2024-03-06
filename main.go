package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

func respondWithPlaceholder(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<h1>Welcome to Switchman</h1> <p>If you see this message, it means that the Switchman web server is running</p>")
}

func respondWith404(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>404</h1> <p>The page you requested does not seem to exist</p>")
}

// ServerConfig is a host container for endpoints
type ServerConfig struct {
	endpoints []Endpoint
}

func (s *ServerConfig) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path
	for _, e := range s.endpoints {
		if strings.HasPrefix(path, e.location) {
			e.handle(w, r)
			return
		}
	}
	respondWith404(w, r)
}

// An endpoint is the main unit of routing inside the server. Incoming requests
// are handled by corresponding EndpointFunction
type Endpoint struct {
	location string
	function EndpointFunction
}

func (ep *Endpoint) handle(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path
	// TODO: customize trimming
	var localPath = strings.TrimPrefix(path, ep.location)

	ep.function.Serve(w, r, localPath)
}

// An endpoint function provides the action that will be applied to requests
// received by its parent endpoint
type EndpointFunction interface {
	Serve(w http.ResponseWriter, r *http.Request, localPath string)
}

// EndpointFiles is an endpoint function that serves files from local filesystem
type EndpointFiles struct {
	fileRoot string // Path to the directory that the files will be served from
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
	var fpath = filepath.Join(f.fileRoot, localPath)
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
	target string
}

func (f *EndpointRedirect) Serve(w http.ResponseWriter, r *http.Request, localPath string) {
	http.Redirect(w, r, f.target, http.StatusMovedPermanently)
}

func readConfig(path string) *ServerConfig {
	var file, err = os.Open(path)
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}
	reader, err := NewConfigReader(file)
	if err != nil {
		log.Fatal(err)
	}

	var server = &ServerConfig{}

	reader.ReadExactToken("server")
	reader.ReadExactToken("{")
	fmt.Println("parse: Server block open")
	for end := false; !end; {
		var token = reader.ReadToken()
		switch token {
		case "locations":
			reader.ReadExactToken("{")
			fmt.Println("parse: Locations block open")
			for {
				var path = reader.ReadToken()
				if path == "}" {
					break
				}
				var ep = Endpoint{}
				var special = []string{"{", "}", ":", EOF}
				if slices.Contains(special, path) {
					log.Fatalf("Unexpected %s, location path expected", TokenName(path))
				}
				fmt.Printf("parse: Location path='%s'\n", path)
				ep.location = path
				reader.ReadExactToken(":")
				var ep_type = reader.ReadToken()
				if slices.Contains(special, ep_type) {
					log.Fatalf("Unexpected %s, endpoint type expected", TokenName(ep_type))
				}
				switch ep_type {
				case "files":
					reader.ReadExactToken("{")
					fmt.Println("parse: Files block open")

					reader.ReadExactToken("sources")
					reader.ReadExactToken(":")
					var files_path = reader.ReadToken()
					if slices.Contains(special, files_path) {
						log.Fatalf("Unexpected %s, path expected", TokenName(files_path))
					}
					fmt.Printf("parse: files sources='%s'\n", files_path)
					ep.function = &EndpointFiles{
						fileRoot: files_path,
					}

					reader.ReadExactToken("}")
					fmt.Println("parse: Files block close")
				case "redirect":
					reader.ReadExactToken("{")
					fmt.Println("parse: Redirect block open")

					reader.ReadExactToken("target")
					reader.ReadExactToken(":")

					var target = reader.ReadToken()
					if slices.Contains(special, target) {
						log.Fatalf("Unexpected '%s', redirect target expected", TokenName(target))
					}
					fmt.Printf("parse: redirect target='%s'\n", target)
					ep.function = &EndpointRedirect{
						target: target,
					}

					reader.ReadExactToken("}")
					fmt.Println("parse: Redirect block close")
				default:
					log.Fatalf("'%s' is not a recognized endpoint type", ep_type)
				}
				server.endpoints = append(server.endpoints, ep)
			}
			fmt.Println("parse: Locations block close")
		case "}":
			end = true
		default:
			log.Fatalf("Unexpected %s, 'locations' or '}' expected", TokenName(token))
		}
	}
	fmt.Println("parse: Server block close")
	reader.ReadExactToken(EOF)
	return server
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing config file argument")
	}
	var config_path = os.Args[1]
	var srv = readConfig(config_path)

	fmt.Println("Switchman web server starting up")
	log.Fatal(http.ListenAndServe(":8080", srv))
}
