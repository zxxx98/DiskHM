package discovery

import (
	"context"
	"io/fs"

	"github.com/example/diskhm/internal/domain"
)

type Service interface {
	Snapshot(ctx context.Context) (domain.DiscoverySnapshot, error)
}

type service struct {
	fsys fs.FS
}

func NewService(fsys fs.FS) Service {
	return service{fsys: fsys}
}

func (s service) Snapshot(ctx context.Context) (domain.DiscoverySnapshot, error) {
	_ = ctx

	disks, err := readDisks(s.fsys)
	if err != nil {
		return domain.DiscoverySnapshot{}, err
	}

	mounts, err := readMounts(s.fsys, disks)
	if err != nil {
		return domain.DiscoverySnapshot{}, err
	}

	return domain.DiscoverySnapshot{Disks: disks, Mounts: mounts}, nil
}
