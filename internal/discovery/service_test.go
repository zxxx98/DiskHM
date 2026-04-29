package discovery

import (
	"context"
	"os"
	"testing"
	"testing/fstest"
)

func TestDiscoverSnapshotFromFixtures(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"sys/block/sda/queue/rotational": &fstest.MapFile{Data: []byte("1\n")},
		"sys/block/sda/size":             &fstest.MapFile{Data: []byte("3907029168\n")},
		"proc/self/mountinfo":            &fstest.MapFile{Data: []byte("29 22 8:1 / /mnt/media rw,relatime - ext4 /dev/sda1 rw\n")},
		"run/udev/data/b8:0":             &fstest.MapFile{Data: []byte("E:ID_MODEL=WD_Red\nE:ID_SERIAL_SHORT=WD-ABC123\n")},
	}

	service := NewService(fsys)
	snapshot, err := service.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	if len(snapshot.Disks) != 1 {
		t.Fatalf("len(snapshot.Disks) = %d, want 1", len(snapshot.Disks))
	}

	disk := snapshot.Disks[0]
	if disk.Name != "sda" {
		t.Fatalf("disk.Name = %q, want %q", disk.Name, "sda")
	}
	if disk.Path != "/dev/sda" {
		t.Fatalf("disk.Path = %q, want %q", disk.Path, "/dev/sda")
	}
	if disk.Model != "WD Red" {
		t.Fatalf("disk.Model = %q, want %q", disk.Model, "WD Red")
	}
	if disk.Serial != "WD-ABC123" {
		t.Fatalf("disk.Serial = %q, want %q", disk.Serial, "WD-ABC123")
	}
	if !disk.Rotational {
		t.Fatal("disk.Rotational = false, want true")
	}
	if disk.SizeBytes != 3907029168*512 {
		t.Fatalf("disk.SizeBytes = %d, want %d", disk.SizeBytes, uint64(3907029168*512))
	}

	if len(snapshot.Mounts) != 1 {
		t.Fatalf("len(snapshot.Mounts) = %d, want 1", len(snapshot.Mounts))
	}

	mount := snapshot.Mounts[0]
	if mount.Source != "/dev/sda1" {
		t.Fatalf("mount.Source = %q, want %q", mount.Source, "/dev/sda1")
	}
	if mount.Target != "/mnt/media" {
		t.Fatalf("mount.Target = %q, want %q", mount.Target, "/mnt/media")
	}
	if mount.DiskID != disk.ID {
		t.Fatalf("mount.DiskID = %q, want %q", mount.DiskID, disk.ID)
	}
}

func TestDiscoverSnapshotFromTestdataFixtures(t *testing.T) {
	t.Parallel()

	service := NewService(os.DirFS("testdata/fixtures/basic"))
	snapshot, err := service.Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	if len(snapshot.Disks) != 1 {
		t.Fatalf("len(snapshot.Disks) = %d, want 1", len(snapshot.Disks))
	}
	if snapshot.Disks[0].Model != "WD Red" {
		t.Fatalf("snapshot.Disks[0].Model = %q, want %q", snapshot.Disks[0].Model, "WD Red")
	}
}
