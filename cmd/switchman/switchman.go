package main

import (
	"log"
	"os"

	"github.com/arrowinaknee/switchman/pkg/api"
	"github.com/arrowinaknee/switchman/pkg/auth"
	"github.com/arrowinaknee/switchman/pkg/runtime"
	"github.com/arrowinaknee/switchman/pkg/settings"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Missing config file argument")
	}
	if len(os.Args) < 3 {
		log.Fatal("Missing settings file argument")
	}
	config_path := os.Args[1]
	settings_path := os.Args[2]

	store := &settings.YamlStore{
		Path: settings_path,
	}

	runtime := runtime.New()
	auth, err := auth.NewManager(store)
	if err != nil {
		log.Fatal(err)
	}

	api.Start(runtime, auth, ":3315")

	err = runtime.LoadServer(config_path)
	if err != nil {
		log.Fatal(err)
	}
	err = runtime.Start()
	if err != nil {
		log.Fatal(err)
	}
}
