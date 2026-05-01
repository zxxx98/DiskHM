package api

import (
	"context"
	"net/http"

	"github.com/example/diskhm/internal/domain"
)

type Dependencies struct {
	TokenPlaintext string
	Runtime        Runtime
}

func NewRouter(deps Dependencies) http.Handler {
	if deps.Runtime == nil {
		deps.Runtime = noopRuntime{}
	}

	mux := http.NewServeMux()
	registerHealthRoutes(mux)
	registerSessionRoutes(mux, deps)
	registerDiskRoutes(mux, deps)
	registerEventRoutes(mux, deps)
	registerTopologyRoutes(mux, deps)
	registerSettingsRoutes(mux, deps)
	registerStaticRoutes(mux)

	return mux
}

type Runtime interface {
	ListDisks(context.Context) ([]domain.DiskView, error)
	Topology(context.Context) (domain.TopologyGraph, error)
	Settings(context.Context) (domain.SettingsView, error)
	ListEvents(context.Context, int) ([]domain.Event, error)
	SubscribeEvents() (<-chan domain.Event, func())
	SleepNow(context.Context, string) error
	SleepAfter(context.Context, string, int) error
	CancelSleep(context.Context, string) error
	RefreshSafe(context.Context, string) error
	RefreshWake(context.Context, string) error
}

type noopRuntime struct{}

func (noopRuntime) ListDisks(context.Context) ([]domain.DiskView, error) {
	return []domain.DiskView{}, nil
}

func (noopRuntime) Topology(context.Context) (domain.TopologyGraph, error) {
	return domain.TopologyGraph{
		Nodes: []domain.TopologyNode{},
		Edges: []domain.TopologyEdge{},
	}, nil
}

func (noopRuntime) Settings(context.Context) (domain.SettingsView, error) {
	return domain.SettingsView{QuietGraceSeconds: 10}, nil
}

func (noopRuntime) ListEvents(context.Context, int) ([]domain.Event, error) {
	return []domain.Event{}, nil
}

func (noopRuntime) SubscribeEvents() (<-chan domain.Event, func()) {
	ch := make(chan domain.Event)
	close(ch)
	return ch, func() {}
}

func (noopRuntime) SleepNow(context.Context, string) error        { return nil }
func (noopRuntime) SleepAfter(context.Context, string, int) error { return nil }
func (noopRuntime) CancelSleep(context.Context, string) error     { return nil }
func (noopRuntime) RefreshSafe(context.Context, string) error     { return nil }
func (noopRuntime) RefreshWake(context.Context, string) error     { return nil }
