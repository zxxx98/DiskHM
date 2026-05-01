package app

import (
	"context"
	"testing"

	"github.com/example/diskhm/internal/config"
	"github.com/example/diskhm/internal/domain"
)

func TestRuntimeBuildsTopologyFromDiscoverySnapshot(t *testing.T) {
	t.Parallel()

	runtime := newTestRuntime(domain.DiscoverySnapshot{
		Disks: []domain.Disk{
			{ID: "disk-sda", Name: "sda", Path: "/dev/sda", Model: "WD Red", Rotational: true},
		},
		Mounts: []domain.Mount{
			{DiskID: "disk-sda", Source: "/dev/sda1", Target: "/srv/data"},
		},
	})

	graph, err := runtime.Topology(context.Background())
	if err != nil {
		t.Fatalf("Topology() error = %v", err)
	}

	if len(graph.Nodes) != 2 {
		t.Fatalf("len(graph.Nodes) = %d, want 2", len(graph.Nodes))
	}
	if len(graph.Edges) != 1 {
		t.Fatalf("len(graph.Edges) = %d, want 1", len(graph.Edges))
	}
	if graph.Edges[0].From != "disk-sda" {
		t.Fatalf("graph.Edges[0].From = %q, want %q", graph.Edges[0].From, "disk-sda")
	}
}

type fakeDiscoveryService struct {
	snapshot domain.DiscoverySnapshot
	err      error
}

func (f fakeDiscoveryService) Snapshot(context.Context) (domain.DiscoverySnapshot, error) {
	return f.snapshot, f.err
}

type fakeEventStore struct{}

func (fakeEventStore) AppendEvent(context.Context, domain.Event) error {
	return nil
}

func (fakeEventStore) ListEvents(context.Context, int) ([]domain.Event, error) {
	return nil, nil
}

func (fakeEventStore) UpsertDisk(context.Context, domain.Disk) error {
	return nil
}

func newTestRuntime(snapshot domain.DiscoverySnapshot) *Runtime {
	return NewRuntime(config.Default(), fakeDiscoveryService{snapshot: snapshot}, fakeEventStore{})
}
