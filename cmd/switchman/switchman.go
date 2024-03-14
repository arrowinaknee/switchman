package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/arrowinaknee/switchman/pkg/appconfig"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing config file argument")
	}
	var config_path = os.Args[1]
	var config_file, err = os.Open(config_path)
	defer config_file.Close()
	if err != nil {
		log.Fatal(err)
	}
	srv, err := appconfig.ParseServer(config_file)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Switchman web server starting up")
	log.Fatal(http.ListenAndServe(":8080", srv))
}
