package sqlitestore

import (
	"context"
	"testing"
	"time"

	"github.com/example/diskhm/internal/domain"
)

func TestStoreUpsertDiskAndAppendEvent(t *testing.T) {
	t.Parallel()

	store, err := Open("file:test-store?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	disk := domain.Disk{
		ID:         "disk-sda",
		Name:       "sda",
		Path:       "/dev/sda",
		Model:      "WD Red",
		Serial:     "WD-123456",
		Transport:  "sata",
		SizeBytes:  uint64(4000787030016),
		Rotational: true,
	}

	if err := store.UpsertDisk(ctx, disk); err != nil {
		t.Fatalf("UpsertDisk() error = %v", err)
	}

	disks, err := store.ListDisks(ctx)
	if err != nil {
		t.Fatalf("ListDisks() error = %v", err)
	}

	if len(disks) != 1 {
		t.Fatalf("len(ListDisks()) = %d, want 1", len(disks))
	}

	if disks[0].ID != "disk-sda" {
		t.Fatalf("ListDisks()[0].ID = %q, want %q", disks[0].ID, "disk-sda")
	}

	if disks[0] != disk {
		t.Fatalf("ListDisks()[0] = %#v, want %#v", disks[0], disk)
	}

	createdAt := time.Date(2026, time.April, 29, 10, 30, 0, 0, time.UTC)
	event := domain.Event{
		DiskID:    "disk-sda",
		Kind:      "sleep_requested",
		Message:   "waiting for disk to become idle",
		CreatedAt: createdAt,
	}
	if err := store.AppendEvent(ctx, event); err != nil {
		t.Fatalf("AppendEvent() error = %v", err)
	}

	var got domain.Event
	if err := store.db.QueryRowContext(
		ctx,
		`SELECT id, disk_id, kind, message, created_at FROM events WHERE disk_id = ?`,
		"disk-sda",
	).Scan(&got.ID, &got.DiskID, &got.Kind, &got.Message, &got.CreatedAt); err != nil {
		t.Fatalf("QueryRowContext() error = %v", err)
	}

	if got.DiskID != event.DiskID || got.Kind != event.Kind || got.Message != event.Message || !got.CreatedAt.Equal(event.CreatedAt) {
		t.Fatalf("stored event = %#v, want fields from %#v", got, event)
	}

	if got.ID == 0 {
		t.Fatalf("stored event ID = %d, want non-zero", got.ID)
	}
}
