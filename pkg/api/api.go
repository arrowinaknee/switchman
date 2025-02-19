package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/arrowinaknee/switchman/pkg/appconfig"
	"github.com/arrowinaknee/switchman/pkg/auth"
	"github.com/arrowinaknee/switchman/pkg/runtime"
	"github.com/rs/cors"
)

type Api struct {
	runtime *runtime.Runtime
	auth    *auth.AuthManager
}

func Start(runtime *runtime.Runtime, auth *auth.AuthManager, address string) {
	api := &Api{
		runtime: runtime,
		auth:    auth,
	}
	mux := http.NewServeMux()

	// FIXME: proper CORS rules if needed when webpage is hosted
	c := cors.AllowAll()
	handler := c.Handler(mux)

	mux.HandleFunc("/config", api.handleConfig)
	mux.HandleFunc("/verify", api.handleVerify)
	mux.HandleFunc("/login", api.handleLogin)
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

func (api *Api) handleLogin(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		Username string `json:"login"`
		Password string `json:"password"`
	}
	switch r.Method {
	case http.MethodPost:
		var rd LoginRequest
		err := json.NewDecoder(r.Body).Decode(&rd)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Unable to parse json: %s", err)
			return
		}
		id, err := api.auth.Users.TrySignIn(rd.Username, rd.Password)
		if errors.Is(err, auth.ErrLoginNotFound) || errors.Is(err, auth.ErrPasswordMismatch) {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Incorrect login or password")
			return
		} else if errors.Is(err, auth.ErrUserDisabled) {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "User is disabled")
			return
		} else if err != nil {
			// FIXME: needs to be logged
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Internal server error")
			return
		}
		tok, err := api.auth.IssueToken(id)
		if err != nil {
			// FIXME: needs to be logged
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Internal server error")
			return
		}
		// write cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    tok,
			SameSite: http.SameSiteStrictMode,
		})
		fmt.Fprint(w, "Authorized successfully")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (api *Api) handleUsers(w http.ResponseWriter, r *http.Request) {
	type user struct {
		Id    string `json:"id"`
		Login string `json:"login"`
	}
	switch r.Method {
	case http.MethodGet:
		ids := api.auth.Users.GetUsersIds()
		users := make([]user, len(ids))
		for i, id := range ids {
			users[i].Id = id
			var err error
			users[i].Login, err = api.auth.Users.GetUserLogin(id)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, "Internal server error")
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(users)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Internal server error")
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
