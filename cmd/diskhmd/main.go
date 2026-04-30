package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/example/diskhm/internal/app"
	"github.com/example/diskhm/internal/config"
)

const defaultConfigPath = "/etc/diskhm/config.yaml"

func main() {
	configPath, err := configPathFromArgs(flag.CommandLine, defaultConfigPath, os.Args[1:])
	if err != nil {
		log.Fatalf("parse flags: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	a := app.New(cfg)

	if err := http.ListenAndServe(cfg.Server.ListenAddr, a.Handler); err != nil {
		log.Fatalf("listen and serve: %v", err)
	}
}

func configPathFromArgs(fs *flag.FlagSet, defaultPath string, args []string) (string, error) {
	configPath := defaultPath
	fs.StringVar(&configPath, "config", defaultPath, "path to config file")

	if err := fs.Parse(args); err != nil {
		return "", err
	}

	return configPath, nil
}
