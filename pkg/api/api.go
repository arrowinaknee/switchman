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

	mux.HandleFunc("GET  /config", api.getConfig)
	mux.HandleFunc("POST /config", api.saveConfig)
	mux.HandleFunc("POST /verify", api.verifyConfig)
	mux.HandleFunc("POST /login", api.login)
	mux.HandleFunc("GET  /users", api.getUserList)
	mux.HandleFunc("POST /users", api.createUser)
	mux.HandleFunc("GET  /users/{login}", api.getUser)
	mux.HandleFunc("POST /users/{login}/login", api.setUserLogin)

	fmt.Printf("api: listening on %s\n", address)
	go http.ListenAndServe(address, handler)
}

func (api *Api) getConfig(w http.ResponseWriter, r *http.Request) {
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
}
func (api *Api) saveConfig(w http.ResponseWriter, r *http.Request) {
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
}

func (api *Api) verifyConfig(w http.ResponseWriter, r *http.Request) {
	_, err := appconfig.ParseServer(r.Body)
	if err != nil {
		fmt.Fprint(w, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (api *Api) login(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		Username string `json:"login"`
		Password string `json:"password"`
	}
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
}

func (api *Api) getUserList(w http.ResponseWriter, r *http.Request) {
	type user struct {
		Id    string `json:"id"`
		Login string `json:"login"`
	}
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
}

func (api *Api) createUser(w http.ResponseWriter, r *http.Request) {
	type user struct {
		Id    string `json:"id"`
		Login string `json:"login"`
	}
	var u struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Unable to parse json: %s", err)
		return
	}
	id, err := api.auth.Users.Create(u.Login, u.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to create user")
	}
	cr := &user{
		Id:    id,
		Login: u.Login,
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(cr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		return
	}
}

func (api *Api) getUser(w http.ResponseWriter, r *http.Request) {
	login := r.PathValue("login")
	if len(login) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "User login empty")
		return
	}

	id, err := api.auth.Users.GetIdByLogin(login)
	if errors.Is(err, auth.ErrLoginNotFound) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "User does not exist")
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		fmt.Printf("Error looking up user id: %v\n", err)
	}

	resp := &struct {
		Id string `json:"id"`
	}{id}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal server error")
		fmt.Printf("Error encoding response: %v\n", err)
	}
}

func (api *Api) setUserLogin(w http.ResponseWriter, r *http.Request) {
	login := r.PathValue("login")

	id, err := api.auth.Users.GetIdByLogin(login)
	if errors.Is(err, auth.ErrLoginNotFound) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "User does not exist")
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal server error")
		fmt.Printf("Error looking up user id: %v\n", err)
	}

	var rdata struct {
		Login string `json:"login"`
	}
	err = json.NewDecoder(r.Body).Decode(&rdata)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Could not decode request json")
		return
	}

	err = api.auth.Users.SetUserLogin(id, rdata.Login)
	if errors.Is(err, auth.ErrLoginEmpty) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid login")
		return
	}
	if errors.Is(err, auth.ErrLoginExists) {
		w.WriteHeader(http.StatusConflict)
		fmt.Fprintf(w, "Login already taken")
		return
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Could not update login")
		fmt.Printf("Error")
		return
	}

	resp := &struct {
		Id string `json:"id"`
	}{id}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Internal server error")
		fmt.Printf("Error encoding response: %v\n", err)
	}
}
