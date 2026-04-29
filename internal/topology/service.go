package topology

import (
	"github.com/example/diskhm/internal/domain"
)

func Build(disks []domain.Disk, mounts []domain.Mount) domain.TopologyGraph {
	nodes := make([]domain.TopologyNode, 0, len(disks)+len(mounts))
	edges := make([]domain.TopologyEdge, 0, len(mounts))

	for _, disk := range disks {
		label := disk.Path
		if disk.Model != "" {
			label = disk.Model
		}
		nodes = append(nodes, domain.TopologyNode{
			ID:    disk.ID,
			Kind:  "disk",
			Label: label,
		})
	}

	for _, mount := range mounts {
		mountID := "mount-" + mount.Target
		nodes = append(nodes, domain.TopologyNode{
			ID:    mountID,
			Kind:  "mount",
			Label: mount.Target,
		})
		if mount.DiskID != "" {
			edges = append(edges, domain.TopologyEdge{From: mount.DiskID, To: mountID})
		}
	}

	return domain.TopologyGraph{Nodes: nodes, Edges: edges}
}
