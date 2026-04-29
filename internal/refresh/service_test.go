package refresh

import (
	"context"
	"errors"
	"testing"

	"github.com/example/diskhm/internal/domain"
)

func TestSafeRefreshCallsSafe(t *testing.T) {
	t.Parallel()

	disk := domain.Disk{ID: "disk-sda", Path: "/dev/sda"}
	probe := &FakeProbe{}
	service := NewService(probe)

	if err := service.SafeRefresh(context.Background(), disk); err != nil {
		t.Fatalf("SafeRefresh() error = %v", err)
	}

	if probe.SafeCalls != 1 {
		t.Fatalf("probe.SafeCalls = %d, want 1", probe.SafeCalls)
	}

	if probe.WakeCalls != 0 {
		t.Fatalf("probe.WakeCalls = %d, want 0", probe.WakeCalls)
	}
}

func TestWakeRefreshCallsWake(t *testing.T) {
	t.Parallel()

	disk := domain.Disk{ID: "disk-sda", Path: "/dev/sda"}
	probe := &FakeProbe{}
	service := NewService(probe)

	if err := service.WakeRefresh(context.Background(), disk); err != nil {
		t.Fatalf("WakeRefresh() error = %v", err)
	}

	if probe.WakeCalls != 1 {
		t.Fatalf("probe.WakeCalls = %d, want 1", probe.WakeCalls)
	}

	if probe.SafeCalls != 0 {
		t.Fatalf("probe.SafeCalls = %d, want 0", probe.SafeCalls)
	}
}

func TestSafeRefreshPropagatesError(t *testing.T) {
	t.Parallel()

	disk := domain.Disk{ID: "disk-sda", Path: "/dev/sda"}
	wantErr := errors.New("safe failed")
	probe := &FakeProbe{SafeErr: wantErr}
	service := NewService(probe)

	err := service.SafeRefresh(context.Background(), disk)
	if !errors.Is(err, wantErr) {
		t.Fatalf("SafeRefresh() error = %v, want %v", err, wantErr)
	}
}

func TestWakeRefreshPropagatesError(t *testing.T) {
	t.Parallel()

	disk := domain.Disk{ID: "disk-sda", Path: "/dev/sda"}
	wantErr := errors.New("wake failed")
	probe := &FakeProbe{WakeErr: wantErr}
	service := NewService(probe)

	err := service.WakeRefresh(context.Background(), disk)
	if !errors.Is(err, wantErr) {
		t.Fatalf("WakeRefresh() error = %v, want %v", err, wantErr)
	}
}

type FakeProbe struct {
	SafeCalls int
	WakeCalls int
	SafeErr   error
	WakeErr   error
}

func (p *FakeProbe) Safe(context.Context, domain.Disk) error {
	p.SafeCalls++
	return p.SafeErr
}

func (p *FakeProbe) Wake(context.Context, domain.Disk) error {
	p.WakeCalls++
	return p.WakeErr
}
