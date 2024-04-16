package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/hello/", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		who := q.Get("who")
		if who == "" {
			who = "someone"
		}
		fmt.Fprintf(w, "Hello, %s!", who)
	})

	log.Fatal(http.ListenAndServe(":6060", nil))
}
