package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//lint:ignore U1000 Ignore unused function
func respondWithPlaceholder(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<h1>Welcome to Switchman</h1> <p>If you see this message, it means that the Switchman web server is running</p>")
}

func respondWith404(w http.ResponseWriter, _ *http.Request) {
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
	tokens, err := NewTokenReader(file)
	if err != nil {
		log.Fatal(err)
	}

	var server = &ServerConfig{}

	tokens.ReadExact("server")
	tokens.ReadExact("{")
	fmt.Println("parse: Server block open")
	for end := false; !end; {
		var token = tokens.ReadNext()
		switch token {
		case "locations":
			tokens.ReadExact("{")
			fmt.Println("parse: Locations block open")
			for {
				var path = tokens.ReadNext()
				if path == "}" {
					break
				}
				var ep = Endpoint{}
				if !path.IsLiteral() {
					log.Fatalf("Unexpected %s, location path expected", path.Quote())
				}
				fmt.Printf("parse: Location path='%s'\n", path)
				ep.location = path.String()
				tokens.ReadExact(":")
				var ep_type = tokens.ReadLiteral()
				switch ep_type {
				case "files":
					tokens.ReadExact("{")
					fmt.Println("parse: Files block open")

					tokens.ReadExact("sources")
					tokens.ReadExact(":")
					var files_path = tokens.ReadLiteral()
					fmt.Printf("parse: files sources='%s'\n", files_path)
					ep.function = &EndpointFiles{
						fileRoot: files_path.String(),
					}

					tokens.ReadExact("}")
					fmt.Println("parse: Files block close")
				case "redirect":
					tokens.ReadExact("{")
					fmt.Println("parse: Redirect block open")

					tokens.ReadExact("target")
					tokens.ReadExact(":")

					var target = tokens.ReadLiteral()
					fmt.Printf("parse: redirect target='%s'\n", target)
					ep.function = &EndpointRedirect{
						target: target.String(),
					}

					tokens.ReadExact("}")
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
			log.Fatalf("Unexpected %s, 'locations' or '}' expected", token.Quote())
		}
	}
	fmt.Println("parse: Server block close")
	tokens.ReadExact(EOF)
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
