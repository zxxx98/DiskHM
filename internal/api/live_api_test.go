package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/diskhm/internal/domain"
)

func TestDisksEndpointReturnsDiscoveredDisk(t *testing.T) {
	t.Parallel()

	runtime := &stubRuntime{
		disks: []domain.DiskView{
			{
				ID:               "disk-sda",
				Name:             "sda",
				Path:             "/dev/sda",
				Model:            "WD Red",
				PowerState:       "unknown",
				RefreshFreshness: "cached",
				Unsupported:      false,
			},
		},
	}

	router := NewRouter(Dependencies{
		TokenPlaintext: "dev-token",
		Runtime:        runtime,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/disks", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionCookieValue})
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), `"disk-sda"`) {
		t.Fatalf("body = %q, want disk payload", rec.Body.String())
	}
}

func TestSleepNowEndpointRunsRuntimeAction(t *testing.T) {
	t.Parallel()

	runtime := &stubRuntime{}
	router := NewRouter(Dependencies{
		TokenPlaintext: "dev-token",
		Runtime:        runtime,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/disks/disk-sda/sleep-now", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionCookieValue})
	req.Header.Set("X-CSRF-Token", csrfTokenValue)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if runtime.sleepNowDiskID != "disk-sda" {
		t.Fatalf("sleepNowDiskID = %q, want %q", runtime.sleepNowDiskID, "disk-sda")
	}
}

type stubRuntime struct {
	disks          []domain.DiskView
	topology       domain.TopologyGraph
	settings       domain.SettingsView
	events         []domain.Event
	sleepNowDiskID string
}

func (s *stubRuntime) ListDisks(context.Context) ([]domain.DiskView, error) {
	return s.disks, nil
}

func (s *stubRuntime) Topology(context.Context) (domain.TopologyGraph, error) {
	return s.topology, nil
}

func (s *stubRuntime) Settings(context.Context) (domain.SettingsView, error) {
	if s.settings.QuietGraceSeconds == 0 {
		return domain.SettingsView{QuietGraceSeconds: 10}, nil
	}
	return s.settings, nil
}

func (s *stubRuntime) ListEvents(context.Context, int) ([]domain.Event, error) {
	return s.events, nil
}

func (s *stubRuntime) SubscribeEvents() (<-chan domain.Event, func()) {
	ch := make(chan domain.Event)
	close(ch)
	return ch, func() {}
}

func (s *stubRuntime) SleepNow(_ context.Context, diskID string) error {
	s.sleepNowDiskID = diskID
	return nil
}

func (s *stubRuntime) SleepAfter(context.Context, string, int) error {
	return nil
}

func (s *stubRuntime) CancelSleep(context.Context, string) error {
	return nil
}

func (s *stubRuntime) RefreshSafe(context.Context, string) error {
	return nil
}

func (s *stubRuntime) RefreshWake(context.Context, string) error {
	return nil
}
