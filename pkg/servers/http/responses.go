package http

import (
	"fmt"
	"net/http"
)

func respondWith404(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "<h1>404</h1> <p>The page you requested does not seem to exist</p>")
}
