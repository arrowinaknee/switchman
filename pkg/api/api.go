package api

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/arrowinaknee/switchman/pkg/appconfig"
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
	mux.HandleFunc("/verify", api.handleVerify)
	go http.ListenAndServe(address, handler)
}

func (api *Api) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		path := api.runtime.GetConfigPath()
		file, err := os.Open(path)
		if err != nil {
			log.Printf("api: error reading config file '%s': %s", path, err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Error reading config file")
			return
		}
		_, err = io.Copy(w, file)
		if err != nil {
			log.Printf("api: error reading config file '%s': %s", path, err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Error reading config file")
			return
		}
	case http.MethodPost:
		path := api.runtime.GetConfigPath()
		// body needs to be both parsed and saved to disk
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("api: error updating config file '%s': %s", path, err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Error updating config file")
			return
		}
		// first parse the config, if code is valid first update the file, then update srv in runtime
		srv, err := appconfig.ParseServer(bytes.NewReader(body))
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			fmt.Fprint(w, err.Error())
			return
		}
		file, err := os.Create(path)
		if err != nil {
			log.Printf("api: error updating config file '%s': %s", path, err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Error updating config file")
			return
		}
		_, err = io.Copy(file, bytes.NewReader(body))
		if err != nil {
			log.Printf("api: error updating config file '%s': %s", path, err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Error updating config file")
			return
		}
		api.runtime.UpdateServer(srv)

		w.WriteHeader(http.StatusOK)
		log.Printf("api: updated config file '%s'", path)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (api *Api) handleVerify(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		_, err := appconfig.ParseServer(r.Body)
		if err != nil {
			fmt.Fprint(w, err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
