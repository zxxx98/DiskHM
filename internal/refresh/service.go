package refresh

import (
	"context"

	"github.com/example/diskhm/internal/domain"
)

type Probe interface {
	Safe(context.Context, domain.Disk) error
	Wake(context.Context, domain.Disk) error
}

type Service struct {
	probe Probe
}

func NewService(probe Probe) Service {
	return Service{probe: probe}
}

func (s Service) SafeRefresh(ctx context.Context, disk domain.Disk) error {
	return s.probe.Safe(ctx, disk)
}

func (s Service) WakeRefresh(ctx context.Context, disk domain.Disk) error {
	return s.probe.Wake(ctx, disk)
}
