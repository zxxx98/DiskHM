package refresh

import (
	"context"
	"testing"

	"github.com/example/diskhm/internal/domain"
)

func TestSafeRefreshSkipsWakeCapableProbe(t *testing.T) {
	t.Parallel()

	disk := domain.Disk{ID: "disk-sda", Path: "/dev/sda"}
	probe := &FakeProbe{}
	service := NewService(probe)

	if err := service.SafeRefresh(context.Background(), disk); err != nil {
		t.Fatalf("SafeRefresh() error = %v", err)
	}

	if probe.WakeCalls != 0 {
		t.Fatalf("probe.WakeCalls = %d, want 0", probe.WakeCalls)
	}
}

type FakeProbe struct {
	SafeCalls int
	WakeCalls int
}

func (p *FakeProbe) Safe(context.Context, domain.Disk) error {
	p.SafeCalls++
	return nil
}

func (p *FakeProbe) Wake(context.Context, domain.Disk) error {
	p.WakeCalls++
	return nil
}
