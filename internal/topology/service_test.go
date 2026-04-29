package topology

import (
	"testing"

	"github.com/example/diskhm/internal/domain"
)

func TestBuildTopologyAddsDiskAndMountNodes(t *testing.T) {
	t.Parallel()

	disks := []domain.Disk{{ID: "disk-sda", Name: "sda", Path: "/dev/sda"}}
	mounts := []domain.Mount{{Source: "/dev/sda1", Target: "/mnt/media", DiskID: "disk-sda"}}

	graph := Build(disks, mounts)

	if len(graph.Nodes) < 2 {
		t.Fatalf("len(graph.Nodes) = %d, want at least 2", len(graph.Nodes))
	}

	if len(graph.Edges) != 1 {
		t.Fatalf("len(graph.Edges) = %d, want 1", len(graph.Edges))
	}
}
