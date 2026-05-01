package app

import (
	"net/http"
	"os"
	"time"

	"github.com/example/diskhm/internal/api"
	"github.com/example/diskhm/internal/config"
	"github.com/example/diskhm/internal/discovery"
	"github.com/example/diskhm/internal/power"
	"github.com/example/diskhm/internal/refresh"
	"github.com/example/diskhm/internal/scheduler"
	sqlitestore "github.com/example/diskhm/internal/store/sqlite"
)

type App struct {
	Handler http.Handler
	Config  config.Config
}

func New(cfg config.Config) *App {
	runtime := NewRuntime(cfg, discovery.NewService(os.DirFS("/")), openEventStore())
	runtime.SetRefreshService(refresh.NewService(noopProbe{}))
	runtime.SetSchedulerService(scheduler.NewService(
		power.Executor{},
		sysfsQuietSampler{root: "/sys/block", window: time.Duration(cfg.Sleep.QuietGraceSeconds) * time.Second},
		syncFlusher{},
	))

	return &App{
		Handler: api.NewRouter(api.Dependencies{
			TokenPlaintext: cfg.Security.TokenHash,
			Runtime:        runtime,
		}),
		Config: cfg,
	}
}

func openEventStore() EventStore {
	store, err := sqlitestore.Open("/var/lib/diskhm/diskhm.db")
	if err != nil {
		return nil
	}
	return store
}
