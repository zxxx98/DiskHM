package sqlitestore_test

import (
	"context"
	"testing"

	"github.com/example/diskhm/internal/domain"
	sqlitestore "github.com/example/diskhm/internal/store/sqlite"
)

func TestStoreUpsertDiskAndAppendEvent(t *testing.T) {
	t.Parallel()

	store, err := sqlitestore.Open("file:test-store?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	disk := domain.Disk{
		ID:    "disk-sda",
		Name:  "sda",
		Model: "WD Red",
		Path:  "/dev/sda",
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

	if err := store.AppendEvent(ctx, domain.Event{
		DiskID: "disk-sda",
		Kind:   "sleep_requested",
	}); err != nil {
		t.Fatalf("AppendEvent() error = %v", err)
	}
}
