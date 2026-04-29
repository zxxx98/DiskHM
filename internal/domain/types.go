package domain

import "time"

type Disk struct {
	ID         string
	Name       string
	Path       string
	Model      string
	Serial     string
	Transport  string
	SizeBytes  uint64
	Rotational bool
}

type Event struct {
	ID        int64
	DiskID    string
	Kind      string
	Message   string
	CreatedAt time.Time
}

type Mount struct {
	Source string
	Target string
	DiskID string
}

type DiscoverySnapshot struct {
	Disks  []Disk
	Mounts []Mount
}

type TopologyNode struct {
	ID    string
	Kind  string
	Label string
}

type TopologyEdge struct {
	From string
	To   string
}

type TopologyGraph struct {
	Nodes []TopologyNode
	Edges []TopologyEdge
}
