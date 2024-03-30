package main

import (
	"log"
	"os"

	"github.com/arrowinaknee/switchman/pkg/api"
	"github.com/arrowinaknee/switchman/pkg/runtime"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing config file argument")
	}
	config_path := os.Args[1]

	runtime := runtime.New()

	api.Start(runtime, ":3315")

	err := runtime.LoadServer(config_path)
	if err != nil {
		log.Fatal(err)
	}
	err = runtime.Start()
	if err != nil {
		log.Fatal(err)
	}
}
