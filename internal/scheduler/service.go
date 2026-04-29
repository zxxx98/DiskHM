package scheduler

import (
	"context"
	"errors"

	"github.com/example/diskhm/internal/domain"
)

var ErrDiskBusy = errors.New("disk is busy")

type Executor interface {
	Standby(context.Context, string) error
}

type Sampler interface {
	IsQuiet(context.Context, domain.Disk) (bool, error)
}

type Flusher interface {
	Flush(context.Context, domain.Disk) error
}

type Service struct {
	executor Executor
	sampler  Sampler
	flusher  Flusher
}

func NewService(executor Executor, sampler Sampler, flusher Flusher) Service {
	return Service{
		executor: executor,
		sampler:  sampler,
		flusher:  flusher,
	}
}

func (s Service) RunSleepNow(ctx context.Context, disk domain.Disk) error {
	if err := s.flusher.Flush(ctx, disk); err != nil {
		return err
	}

	quiet, err := s.sampler.IsQuiet(ctx, disk)
	if err != nil {
		return err
	}

	if !quiet {
		return ErrDiskBusy
	}

	return s.executor.Standby(ctx, disk.Path)
}
