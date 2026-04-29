package domain

import "time"

type Disk struct {
	ID         string
	Name       string
	Path       string
	Model      string
	Serial     string
	Transport  string
	SizeBytes  int64
	Rotational bool
}

type Event struct {
	ID        int64
	DiskID    string
	Kind      string
	Message   string
	CreatedAt time.Time
}
