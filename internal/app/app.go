package app

import (
	"net/http"

	"github.com/example/diskhm/internal/api"
	"github.com/example/diskhm/internal/config"
)

type App struct {
	Handler http.Handler
	Config  config.Config
}

func New(cfg config.Config) *App {
	return &App{
		Handler: api.NewRouter(api.Dependencies{}),
		Config:  cfg,
	}
}
