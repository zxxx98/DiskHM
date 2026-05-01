package discovery

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/example/diskhm/internal/domain"
)

func readDisks(fsys fs.FS) ([]domain.Disk, error) {
	entries, err := fs.ReadDir(fsys, "sys/block")
	if err != nil {
		return nil, err
	}

	disks := make([]domain.Disk, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if shouldSkipBlockDevice(name) {
			continue
		}

		rotational, err := readRotational(fsys, name)
		if err != nil {
			return nil, err
		}

		sizeBytes, err := readSizeBytes(fsys, name)
		if err != nil {
			return nil, err
		}

		info, err := readUdevInfo(fsys, name)
		if err != nil {
			return nil, err
		}

		disks = append(disks, domain.Disk{
			ID:         "disk-" + name,
			Name:       name,
			Path:       filepath.ToSlash(filepath.Join("/dev", name)),
			Model:      info.Model,
			Serial:     info.Serial,
			Transport:  info.Transport,
			SizeBytes:  sizeBytes,
			Rotational: rotational,
		})
	}

	return disks, nil
}

func shouldSkipBlockDevice(name string) bool {
	return strings.HasPrefix(name, "loop") ||
		strings.HasPrefix(name, "ram") ||
		strings.HasPrefix(name, "zram")
}

func readRotational(fsys fs.FS, name string) (bool, error) {
	raw, err := fs.ReadFile(fsys, filepath.ToSlash(filepath.Join("sys/block", name, "queue/rotational")))
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(raw)) == "1", nil
}

func readSizeBytes(fsys fs.FS, name string) (uint64, error) {
	raw, err := fs.ReadFile(fsys, filepath.ToSlash(filepath.Join("sys/block", name, "size")))
	if err != nil {
		return 0, err
	}

	sectors, err := strconv.ParseUint(strings.TrimSpace(string(raw)), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s size: %w", name, err)
	}

	return sectors * 512, nil
}

func readDeviceNumber(fsys fs.FS, name string) (string, error) {
	raw, err := fs.ReadFile(fsys, filepath.ToSlash(filepath.Join("sys/block", name, "dev")))
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(raw)), nil
}
