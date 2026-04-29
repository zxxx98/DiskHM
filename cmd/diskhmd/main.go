package main

import (
	"log"
	"net/http"

	"github.com/example/diskhm/internal/app"
	"github.com/example/diskhm/internal/config"
)

const configPath = "/etc/diskhm/config.yaml"

func main() {
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	a := app.New(cfg)

	if err := http.ListenAndServe(cfg.Server.ListenAddr, a.Handler); err != nil {
		log.Fatalf("listen and serve: %v", err)
	}
}
