package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/arrowinaknee/switchman/pkg/runtime"
	"github.com/rs/cors"
)

type Api struct {
	runtime *runtime.Runtime
}

func Start(runtime *runtime.Runtime, address string) {
	api := &Api{
		runtime: runtime,
	}
	mux := http.NewServeMux()

	// FIXME: proper CORS rules if needed when webpage is hosted
	c := cors.AllowAll()
	handler := c.Handler(mux)

	mux.HandleFunc("/config", api.handleConfig)
	go http.ListenAndServe(address, handler)
}

func (api *Api) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		path := api.runtime.GetConfigPath()
		file, err := os.Open(path)
		if err != nil {
			log.Printf("Error reading config file '%s': %s", path, err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Error reading config file")
			return
		}
		_, err = io.Copy(w, file)
		if err != nil {
			log.Printf("Error reading config file '%s': %s", path, err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Error reading config file")
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
