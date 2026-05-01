package domain

import (
	"errors"
	"time"
)

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
	ID        int64     `json:"id"`
	DiskID    string    `json:"diskId"`
	Kind      string    `json:"kind"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
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
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Label string `json:"label"`
}

type TopologyEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type TopologyGraph struct {
	Nodes []TopologyNode `json:"nodes"`
	Edges []TopologyEdge `json:"edges"`
}

type DiskTaskView struct {
	Kind      string    `json:"kind"`
	State     string    `json:"state"`
	ExecuteAt time.Time `json:"executeAt,omitempty"`
	LastError string    `json:"lastError,omitempty"`
}

type DiskView struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	Path             string        `json:"path"`
	Model            string        `json:"model"`
	PowerState       string        `json:"powerState"`
	RefreshFreshness string        `json:"refreshFreshness"`
	Unsupported      bool          `json:"unsupported"`
	Mounts           []string      `json:"mounts,omitempty"`
	Task             *DiskTaskView `json:"task,omitempty"`
}

type SettingsView struct {
	QuietGraceSeconds int `json:"quiet_grace_seconds"`
}

var (
	ErrDiskNotFound      = errors.New("disk not found")
	ErrTaskConflict      = errors.New("task already exists for disk")
	ErrUnsupportedDevice = errors.New("unsupported device")
)
