package scheduler

import (
	"context"
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

	if executor.StandbyCalls != 1 {
		t.Fatalf("executor.StandbyCalls = %d, want 1", executor.StandbyCalls)
	}
}

type FakeExecutor struct {
	StandbyCalls int
	CheckCalls   int
}

func (f *FakeExecutor) Standby(context.Context, string) error {
	f.StandbyCalls++
	return nil
}

func (f *FakeExecutor) CheckState(context.Context, string) error {
	f.CheckCalls++
	return nil
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
}

func (f *FakeFlusher) Flush(context.Context, domain.Disk) error {
	f.Calls++
	return nil
}
