package sqlitestore

import (
	"context"
	"math"
	"strings"
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
	var gotCreatedAt string
	if err := store.db.QueryRowContext(
		ctx,
		`SELECT id, disk_id, kind, message, created_at FROM events WHERE disk_id = ?`,
		"disk-sda",
	).Scan(&got.ID, &got.DiskID, &got.Kind, &got.Message, &gotCreatedAt); err != nil {
		t.Fatalf("QueryRowContext() error = %v", err)
	}

	if got.DiskID != event.DiskID || got.Kind != event.Kind || got.Message != event.Message {
		t.Fatalf("stored event = %#v, want fields from %#v", got, event)
	}

	if got.ID == 0 {
		t.Fatalf("stored event ID = %d, want non-zero", got.ID)
	}

	if gotCreatedAt == "" {
		t.Fatalf("stored event created_at is empty")
	}

	if !strings.Contains(gotCreatedAt, "2026-04-29") {
		t.Fatalf("stored event created_at = %q, want it to include %q", gotCreatedAt, "2026-04-29")
	}
}

func TestMigrationDefaultsMatchTaskPlan(t *testing.T) {
	t.Parallel()

	store, err := Open("file:test-store-defaults?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	assertColumnDefault(t, store, "disks", "serial", "''")
	assertColumnDefault(t, store, "disks", "transport", "''")
	assertColumnDefault(t, store, "disks", "size_bytes", "0")
	assertColumnDefault(t, store, "disks", "rotational", "1")
	assertColumnDefault(t, store, "events", "message", "''")
	assertColumnDefault(t, store, "events", "created_at", "CURRENT_TIMESTAMP")
}

func TestAppendEventFailsForUnknownDisk(t *testing.T) {
	t.Parallel()

	store, err := Open("file:test-store-fk?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	err = store.AppendEvent(context.Background(), domain.Event{
		DiskID:  "missing-disk",
		Kind:    "sleep_requested",
		Message: "should fail",
	})
	if err == nil {
		t.Fatal("AppendEvent() error = nil, want foreign key failure")
	}
}

func TestUpsertDiskRejectsSizeBytesAboveMaxInt64(t *testing.T) {
	t.Parallel()

	store, err := Open("file:test-store-size?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	err = store.UpsertDisk(context.Background(), domain.Disk{
		ID:        "disk-big",
		Name:      "big",
		Path:      "/dev/big",
		Model:     "Big Disk",
		SizeBytes: uint64(math.MaxInt64) + 1,
	})
	if err == nil {
		t.Fatal("UpsertDisk() error = nil, want size range failure")
	}
}

func TestAppendEventWithZeroCreatedAtUsesDatabaseDefault(t *testing.T) {
	t.Parallel()

	store, err := Open("file:test-store-created-at?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.UpsertDisk(ctx, domain.Disk{
		ID:   "disk-sda",
		Name: "sda",
		Path: "/dev/sda",
	}); err != nil {
		t.Fatalf("UpsertDisk() error = %v", err)
	}

	if err := store.AppendEvent(ctx, domain.Event{
		DiskID:  "disk-sda",
		Kind:    "sleep_requested",
		Message: "timestamp from db",
	}); err != nil {
		t.Fatalf("AppendEvent() error = %v", err)
	}

	var gotCreatedAt string
	if err := store.db.QueryRowContext(
		ctx,
		`SELECT created_at FROM events WHERE disk_id = ?`,
		"disk-sda",
	).Scan(&gotCreatedAt); err != nil {
		t.Fatalf("QueryRowContext() error = %v", err)
	}

	if gotCreatedAt == "" {
		t.Fatal("stored event created_at is empty")
	}

	if strings.Contains(gotCreatedAt, "0001-01-01") {
		t.Fatalf("stored event created_at = %q, want database default timestamp", gotCreatedAt)
	}
}

func TestStoreListEventsReturnsNewestFirst(t *testing.T) {
	t.Parallel()

	store, err := Open("file:test-store-list-events?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.UpsertDisk(ctx, domain.Disk{
		ID:   "disk-sda",
		Name: "sda",
		Path: "/dev/sda",
	}); err != nil {
		t.Fatalf("UpsertDisk() error = %v", err)
	}

	older := time.Date(2026, time.May, 1, 10, 0, 0, 0, time.UTC)
	newer := time.Date(2026, time.May, 1, 11, 0, 0, 0, time.UTC)
	if err := store.AppendEvent(ctx, domain.Event{
		DiskID:    "disk-sda",
		Kind:      "older",
		Message:   "older",
		CreatedAt: older,
	}); err != nil {
		t.Fatalf("AppendEvent() older error = %v", err)
	}
	if err := store.AppendEvent(ctx, domain.Event{
		DiskID:    "disk-sda",
		Kind:      "newer",
		Message:   "newer",
		CreatedAt: newer,
	}); err != nil {
		t.Fatalf("AppendEvent() newer error = %v", err)
	}

	events, err := store.ListEvents(ctx, 10)
	if err != nil {
		t.Fatalf("ListEvents() error = %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("len(events) = %d, want 2", len(events))
	}
	if events[0].Kind != "newer" {
		t.Fatalf("events[0].Kind = %q, want %q", events[0].Kind, "newer")
	}
	if events[1].Kind != "older" {
		t.Fatalf("events[1].Kind = %q, want %q", events[1].Kind, "older")
	}
}

func assertColumnDefault(t *testing.T, store *Store, table string, column string, want string) {
	t.Helper()

	rows, err := store.db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		t.Fatalf("PRAGMA table_info(%s) error = %v", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			t.Fatalf("Scan() error = %v", err)
		}
		if name == column {
			got, _ := defaultValue.(string)
			if got != want {
				t.Fatalf("%s.%s default = %q, want %q", table, column, got, want)
			}
			return
		}
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err() = %v", err)
	}

	t.Fatalf("column %s.%s not found", table, column)
}
