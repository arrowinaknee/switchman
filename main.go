package main

import (
	"bufio"
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
	var reader = bufio.NewReader(file)
	if err != nil {
		log.Fatal(err)
	}

	var tokens = []string{}
	var tok = strings.Builder{}
	for {
		var r, _, err = reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				if tok.Len() > 0 {
					tokens = append(tokens, tok.String())
					tok.Reset()
				}
				break
			} else {
				log.Fatal(err)
			}
		}

		// whitespace ends any token that was being accumulated
		var whitespace = []rune{' ', '\t', '\n', '\r'}
		if slices.Contains(whitespace, r) {
			if tok.Len() > 0 {
				tokens = append(tokens, tok.String())
				tok.Reset()
			}
			continue
		}

		// check for special characters
		var special = []rune{'{', '}', ':'}
		if slices.Contains(special, r) {
			if tok.Len() > 0 {
				tokens = append(tokens, tok.String())
				tok.Reset()
			}
			tok.WriteRune(r)
			tokens = append(tokens, tok.String())
			tok.Reset()
			continue
		}

		// build normal token
		tok.WriteRune(r)
	}

	var server = &ServerConfig{}

	if len(tokens) < 1 {
		log.Fatal("Unexpected EOF, 'server' expected")
	}
	if tokens[0] != "server" {
		log.Fatalf("Unexpected '%s', 'server' expected", tokens[0])
	}
	tokens = tokens[1:]
	if len(tokens) < 1 {
		log.Fatal("Unexpected EOF, '{' expected")
	}
	if tokens[0] != "{" {
		log.Fatalf("Unexpected '%s', '{' expected", tokens[0])
	}
	tokens = tokens[1:]
	fmt.Println("parse: Server block open")
	for end := false; !end; {
		if len(tokens) < 1 {
			log.Fatal("Unexpected EOF, 'locations' or '}' expected")
		}
		var token = tokens[0]
		tokens = tokens[1:]
		switch token {
		case "locations":
			if len(tokens) < 1 {
				log.Fatal("Unexpected EOF, '{' expected")
			}
			if tokens[0] != "{" {
				log.Fatalf("Unexpected '%s', '{' expected", tokens[0])
			}
			tokens = tokens[1:]
			fmt.Println("parse: Locations block open")
			for {
				if len(tokens) < 1 {
					log.Fatal("Unexpected EOF, location path or '}' expected")
				}
				if tokens[0] == "}" {
					break
				}
				var path = tokens[0]
				var ep = Endpoint{}
				var special = []string{"{", "}", ":"}
				if slices.Contains(special, path) {
					log.Fatalf("Unexpected '%s', location path expected", path)
				}
				fmt.Printf("parse: Location path='%s'\n", path)
				ep.location = path
				tokens = tokens[1:]
				if len(tokens) < 1 {
					log.Fatal("Unexpected EOF, ':' expected")
				}
				if tokens[0] != ":" {
					log.Fatalf("Unexpected '%s', ':' expected", tokens[0])
				}
				tokens = tokens[1:]
				if len(tokens) < 1 {
					log.Fatal("Unexpected EOF, endpoint type expected")
				}
				var ep_type = tokens[0]
				tokens = tokens[1:]
				if slices.Contains(special, ep_type) {
					log.Fatalf("Unexpected '%s', endpoint type expected", ep_type)
				}
				switch ep_type {
				case "files":
					if len(tokens) < 1 {
						log.Fatal("Unexpected EOF, '{' expected")
					}
					if tokens[0] != "{" {
						log.Fatalf("Unexpected '%s', '{' expected", tokens[0])
					}
					tokens = tokens[1:]
					fmt.Println("parse: Files block open")

					if len(tokens) < 1 {
						log.Fatal("Unexpected EOF, 'sources' expected")
					}
					if tokens[0] != "sources" {
						log.Fatalf("Unexpected '%s', 'sources' expected", tokens[0])
					}
					tokens = tokens[1:]

					if len(tokens) < 1 {
						log.Fatal("Unexpected EOF, ':' expected")
					}
					if tokens[0] != ":" {
						log.Fatalf("Unexpected '%s', ':' expected", tokens[0])
					}
					tokens = tokens[1:]

					if len(tokens) < 1 {
						log.Fatal("Unexpected EOF, path expected")
					}
					var files_path = tokens[0]
					tokens = tokens[1:]
					if slices.Contains(special, ep_type) {
						log.Fatalf("Unexpected '%s', path", ep_type)
					}
					fmt.Printf("parse: files sources='%s'\n", files_path)
					ep.function = &EndpointFiles{
						fileRoot: files_path,
					}

					if len(tokens) < 1 {
						log.Fatal("Unexpected EOF, '}' expected")
					}
					if tokens[0] != "}" {
						log.Fatalf("Unexpected '%s', '}' expected", tokens[0])
					}
					tokens = tokens[1:]
					fmt.Println("parse: Files block close")
				case "redirect":
					if len(tokens) < 1 {
						log.Fatal("Unexpected EOF, '{' expected")
					}
					if tokens[0] != "{" {
						log.Fatalf("Unexpected '%s', '{' expected", tokens[0])
					}
					tokens = tokens[1:]
					fmt.Println("parse: Redirect block open")

					if len(tokens) < 1 {
						log.Fatal("Unexpected EOF, 'target' expected")
					}
					if tokens[0] != "target" {
						log.Fatalf("Unexpected '%s', 'target' expected", tokens[0])
					}
					tokens = tokens[1:]

					if len(tokens) < 1 {
						log.Fatal("Unexpected EOF, ':' expected")
					}
					if tokens[0] != ":" {
						log.Fatalf("Unexpected '%s', ':' expected", tokens[0])
					}
					tokens = tokens[1:]

					if len(tokens) < 1 {
						log.Fatal("Unexpected EOF, path expected")
					}
					var target = tokens[0]
					tokens = tokens[1:]
					if slices.Contains(special, ep_type) {
						log.Fatalf("Unexpected '%s', path", ep_type)
					}
					fmt.Printf("parse: redirect target='%s'\n", target)
					ep.function = &EndpointRedirect{
						target: target,
					}

					if len(tokens) < 1 {
						log.Fatal("Unexpected EOF, '}' expected")
					}
					if tokens[0] != "}" {
						log.Fatalf("Unexpected '%s', '}' expected", tokens[0])
					}
					tokens = tokens[1:]
					fmt.Println("parse: Redirect block close")
				default:
					log.Fatalf("'%s' is not a recognized endpoint type", ep_type)
				}
				server.endpoints = append(server.endpoints, ep)
			}
			fmt.Println("parse: Locations block close")
		case "}":
			end = true
		}
	}
	fmt.Println("parse: Server block close")
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
