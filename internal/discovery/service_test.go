package discovery

import (
	"context"
	"io/fs"
	"os"
	"slices"
	"testing"
	"testing/fstest"
	"time"
)

func TestDiscoverSnapshotFromFixtures(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"sys/block/sda/queue/rotational": &fstest.MapFile{Data: []byte("1\n")},
		"sys/block/sda/size":             &fstest.MapFile{Data: []byte("3907029168\n")},
		"sys/block/sda/dev":              &fstest.MapFile{Data: []byte("8:0\n")},
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

func TestDiscoverSnapshotParsesMountinfoOptionalFields(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"sys/block/nvme0n1/queue/rotational": &fstest.MapFile{Data: []byte("0\n")},
		"sys/block/nvme0n1/size":             &fstest.MapFile{Data: []byte("1953525168\n")},
		"sys/block/nvme0n1/dev":              &fstest.MapFile{Data: []byte("259:0\n")},
		"proc/self/mountinfo":                &fstest.MapFile{Data: []byte("32 29 259:1 / /mnt/fast rw,relatime shared:12 master:34 - ext4 /dev/nvme0n1p1 rw\n")},
		"run/udev/data/b259:0":               &fstest.MapFile{Data: []byte("E:ID_MODEL=Fast_SSD\nE:ID_SERIAL_SHORT=FAST-001\n")},
	}

	snapshot, err := NewService(fsys).Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	if len(snapshot.Disks) != 1 {
		t.Fatalf("len(snapshot.Disks) = %d, want 1", len(snapshot.Disks))
	}

	if len(snapshot.Mounts) != 1 {
		t.Fatalf("len(snapshot.Mounts) = %d, want 1", len(snapshot.Mounts))
	}

	if snapshot.Mounts[0].Source != "/dev/nvme0n1p1" {
		t.Fatalf("snapshot.Mounts[0].Source = %q, want %q", snapshot.Mounts[0].Source, "/dev/nvme0n1p1")
	}

	if snapshot.Mounts[0].DiskID != snapshot.Disks[0].ID {
		t.Fatalf("snapshot.Mounts[0].DiskID = %q, want %q", snapshot.Mounts[0].DiskID, snapshot.Disks[0].ID)
	}
}

func TestDiscoverSnapshotUsesPerDiskUdevData(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"sys/block/sda/queue/rotational": &fstest.MapFile{Data: []byte("1\n")},
		"sys/block/sda/size":             &fstest.MapFile{Data: []byte("3907029168\n")},
		"sys/block/sda/dev":              &fstest.MapFile{Data: []byte("8:0\n")},
		"sys/block/sdb/queue/rotational": &fstest.MapFile{Data: []byte("1\n")},
		"sys/block/sdb/size":             &fstest.MapFile{Data: []byte("1953514584\n")},
		"sys/block/sdb/dev":              &fstest.MapFile{Data: []byte("8:16\n")},
		"proc/self/mountinfo":            &fstest.MapFile{Data: []byte("29 22 8:1 / /mnt/media rw,relatime - ext4 /dev/sda1 rw\n30 22 8:17 / /mnt/archive rw,relatime - ext4 /dev/sdb1 rw\n")},
		"run/udev/data/b8:0":             &fstest.MapFile{Data: []byte("E:ID_MODEL=WD_Red\nE:ID_SERIAL_SHORT=WD-ABC123\n")},
		"run/udev/data/b8:16":            &fstest.MapFile{Data: []byte("E:ID_MODEL=Seagate_IronWolf\nE:ID_SERIAL_SHORT=SG-XYZ789\n")},
	}

	snapshot, err := NewService(fsys).Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	if len(snapshot.Disks) != 2 {
		t.Fatalf("len(snapshot.Disks) = %d, want 2", len(snapshot.Disks))
	}

	models := []string{snapshot.Disks[0].Model, snapshot.Disks[1].Model}
	serials := []string{snapshot.Disks[0].Serial, snapshot.Disks[1].Serial}
	slices.Sort(models)
	slices.Sort(serials)

	if !slices.Equal(models, []string{"Seagate IronWolf", "WD Red"}) {
		t.Fatalf("disk models = %#v, want distinct per-disk udev data", models)
	}

	if !slices.Equal(serials, []string{"SG-XYZ789", "WD-ABC123"}) {
		t.Fatalf("disk serials = %#v, want distinct per-disk udev data", serials)
	}
}

func TestDiscoverSnapshotSkipsVirtualBlockDevices(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"sys/block/loop0/queue/rotational": &fstest.MapFile{Data: []byte("0\n")},
		"sys/block/loop0/size":             &fstest.MapFile{Data: []byte("1024\n")},
		"sys/block/loop0/dev":              &fstest.MapFile{Data: []byte("7:0\n")},
		"sys/block/zram0/queue/rotational": &fstest.MapFile{Data: []byte("0\n")},
		"sys/block/zram0/size":             &fstest.MapFile{Data: []byte("1024\n")},
		"sys/block/zram0/dev":              &fstest.MapFile{Data: []byte("254:0\n")},
		"sys/block/sda/queue/rotational":   &fstest.MapFile{Data: []byte("1\n")},
		"sys/block/sda/size":               &fstest.MapFile{Data: []byte("3907029168\n")},
		"sys/block/sda/dev":                &fstest.MapFile{Data: []byte("8:0\n")},
		"proc/self/mountinfo":              &fstest.MapFile{Data: []byte("")},
		"run/udev/data/b8:0":               &fstest.MapFile{Data: []byte("E:ID_MODEL=WD_Red\n")},
	}

	snapshot, err := NewService(fsys).Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	if len(snapshot.Disks) != 1 {
		t.Fatalf("len(snapshot.Disks) = %d, want 1", len(snapshot.Disks))
	}
	if snapshot.Disks[0].Name != "sda" {
		t.Fatalf("snapshot.Disks[0].Name = %q, want %q", snapshot.Disks[0].Name, "sda")
	}
}

func TestDiscoverSnapshotToleratesMissingUdevData(t *testing.T) {
	t.Parallel()

	fsys := fstest.MapFS{
		"sys/block/sda/queue/rotational": &fstest.MapFile{Data: []byte("1\n")},
		"sys/block/sda/size":             &fstest.MapFile{Data: []byte("3907029168\n")},
		"sys/block/sda/dev":              &fstest.MapFile{Data: []byte("8:0\n")},
		"proc/self/mountinfo":            &fstest.MapFile{Data: []byte("")},
	}

	snapshot, err := NewService(fsys).Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	if len(snapshot.Disks) != 1 {
		t.Fatalf("len(snapshot.Disks) = %d, want 1", len(snapshot.Disks))
	}
	if snapshot.Disks[0].Model != "" {
		t.Fatalf("snapshot.Disks[0].Model = %q, want empty model when udev data is missing", snapshot.Disks[0].Model)
	}
}

func TestDiscoverSnapshotAcceptsSymlinkBlockEntries(t *testing.T) {
	t.Parallel()

	fsys := symlinkBlockFS{
		MapFS: fstest.MapFS{
			"sys/block/sda/queue/rotational": &fstest.MapFile{Data: []byte("1\n")},
			"sys/block/sda/size":             &fstest.MapFile{Data: []byte("3907029168\n")},
			"sys/block/sda/dev":              &fstest.MapFile{Data: []byte("8:0\n")},
			"proc/self/mountinfo":            &fstest.MapFile{Data: []byte("")},
			"run/udev/data/b8:0":             &fstest.MapFile{Data: []byte("E:ID_MODEL=WD_Red\n")},
		},
		entries: []fs.DirEntry{
			fakeDirEntry{name: "sda", mode: fs.ModeSymlink},
		},
	}

	snapshot, err := NewService(fsys).Snapshot(context.Background())
	if err != nil {
		t.Fatalf("Snapshot() error = %v", err)
	}

	if len(snapshot.Disks) != 1 {
		t.Fatalf("len(snapshot.Disks) = %d, want 1", len(snapshot.Disks))
	}
	if snapshot.Disks[0].Name != "sda" {
		t.Fatalf("snapshot.Disks[0].Name = %q, want %q", snapshot.Disks[0].Name, "sda")
	}
}

type symlinkBlockFS struct {
	fstest.MapFS
	entries []fs.DirEntry
}

func (s symlinkBlockFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if name == "sys/block" {
		return s.entries, nil
	}
	return fs.ReadDir(s.MapFS, name)
}

type fakeDirEntry struct {
	name string
	mode fs.FileMode
}

func (f fakeDirEntry) Name() string               { return f.name }
func (f fakeDirEntry) IsDir() bool                { return f.mode.IsDir() }
func (f fakeDirEntry) Type() fs.FileMode          { return f.mode }
func (f fakeDirEntry) Info() (fs.FileInfo, error) { return fakeFileInfo{name: f.name, mode: f.mode}, nil }

type fakeFileInfo struct {
	name string
	mode fs.FileMode
}

func (f fakeFileInfo) Name() string       { return f.name }
func (f fakeFileInfo) Size() int64        { return 0 }
func (f fakeFileInfo) Mode() fs.FileMode  { return f.mode }
func (f fakeFileInfo) ModTime() time.Time { return time.Unix(0, 0) }
func (f fakeFileInfo) IsDir() bool        { return f.mode.IsDir() }
func (f fakeFileInfo) Sys() any           { return nil }
