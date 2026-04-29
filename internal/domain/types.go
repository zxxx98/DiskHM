package domain

type Disk struct {
	ID    string
	Name  string
	Model string
	Path  string
}

type Event struct {
	DiskID string
	Kind   string
}
