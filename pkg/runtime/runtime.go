package runtime

import (
	"fmt"
	"net/http"
	"os"

	"github.com/arrowinaknee/switchman/pkg/appconfig"
)

type Server interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type Runtime struct {
	server     Server
	configPath string
}

func New() *Runtime {
	return &Runtime{}
}

// load server configuration at specified path and track the locaion
func (r *Runtime) LoadServer(path string) error {
	config_file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer config_file.Close()

	srv, err := appconfig.ParseServer(config_file)
	if err != nil {
		return err
	}

	r.server = srv
	r.configPath = path
	return nil
}

// Update app state to use new server. Does not change the source file
func (r *Runtime) UpdateServer(s Server) {
	r.server = s
}

func (r *Runtime) GetConfigPath() string {
	return r.configPath
}

func (r *Runtime) Start() error {
	fmt.Println("Switchman web server starting up")
	return http.ListenAndServe(":8080", r)
}

func (r *Runtime) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	r.server.ServeHTTP(w, rq)
}
