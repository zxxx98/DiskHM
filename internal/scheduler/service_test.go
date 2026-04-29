package scheduler

import (
	"context"
	"errors"
	"testing"

	"github.com/example/diskhm/internal/domain"
)

func TestSleepNowWaitsForQuietDiskBeforeExecuting(t *testing.T) {
	t.Parallel()

	disk := domain.Disk{ID: "disk-sda", Path: "/dev/sda"}
	executor := &FakeExecutor{}
	sampler := &FakeSampler{Quiet: true}
	flusher := &FakeFlusher{}
	service := NewService(executor, sampler, flusher)

	if err := service.RunSleepNow(context.Background(), disk); err != nil {
		t.Fatalf("RunSleepNow() error = %v", err)
	}

	if flusher.Calls != 1 {
		t.Fatalf("flusher.Calls = %d, want 1", flusher.Calls)
	}

	if sampler.Calls != 1 {
		t.Fatalf("sampler.Calls = %d, want 1", sampler.Calls)
	}

	if executor.StandbyCalls != 1 {
		t.Fatalf("executor.StandbyCalls = %d, want 1", executor.StandbyCalls)
	}
}

func TestSleepNowReturnsErrDiskBusyWhenDiskIsNotQuiet(t *testing.T) {
	t.Parallel()

	disk := domain.Disk{ID: "disk-sda", Path: "/dev/sda"}
	executor := &FakeExecutor{}
	sampler := &FakeSampler{Quiet: false}
	flusher := &FakeFlusher{}
	service := NewService(executor, sampler, flusher)

	err := service.RunSleepNow(context.Background(), disk)
	if !errors.Is(err, ErrDiskBusy) {
		t.Fatalf("RunSleepNow() error = %v, want %v", err, ErrDiskBusy)
	}

	if executor.StandbyCalls != 0 {
		t.Fatalf("executor.StandbyCalls = %d, want 0", executor.StandbyCalls)
	}
}

func TestSleepNowPropagatesFlushError(t *testing.T) {
	t.Parallel()

	disk := domain.Disk{ID: "disk-sda", Path: "/dev/sda"}
	wantErr := errors.New("flush failed")
	executor := &FakeExecutor{}
	sampler := &FakeSampler{Quiet: true}
	flusher := &FakeFlusher{Err: wantErr}
	service := NewService(executor, sampler, flusher)

	err := service.RunSleepNow(context.Background(), disk)
	if !errors.Is(err, wantErr) {
		t.Fatalf("RunSleepNow() error = %v, want %v", err, wantErr)
	}

	if sampler.Calls != 0 {
		t.Fatalf("sampler.Calls = %d, want 0", sampler.Calls)
	}
}

func TestSleepNowPropagatesStandbyError(t *testing.T) {
	t.Parallel()

	disk := domain.Disk{ID: "disk-sda", Path: "/dev/sda"}
	wantErr := errors.New("standby failed")
	executor := &FakeExecutor{Err: wantErr}
	sampler := &FakeSampler{Quiet: true}
	flusher := &FakeFlusher{}
	service := NewService(executor, sampler, flusher)

	err := service.RunSleepNow(context.Background(), disk)
	if !errors.Is(err, wantErr) {
		t.Fatalf("RunSleepNow() error = %v, want %v", err, wantErr)
	}
}

type FakeExecutor struct {
	StandbyCalls int
	Err          error
}

func (f *FakeExecutor) Standby(context.Context, string) error {
	f.StandbyCalls++
	return f.Err
}

type FakeSampler struct {
	Quiet bool
	Calls int
}

func (f *FakeSampler) IsQuiet(context.Context, domain.Disk) (bool, error) {
	f.Calls++
	return f.Quiet, nil
}

type FakeFlusher struct {
	Calls int
	Err   error
}

func (f *FakeFlusher) Flush(context.Context, domain.Disk) error {
	f.Calls++
	return f.Err
}
