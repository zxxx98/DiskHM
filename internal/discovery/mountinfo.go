package discovery

import (
	"bufio"
	"io/fs"
	"slices"
	"strings"

	"github.com/example/diskhm/internal/domain"
)

func readMounts(fsys fs.FS, disks []domain.Disk) ([]domain.Mount, error) {
	raw, err := fs.ReadFile(fsys, "proc/self/mountinfo")
	if err != nil {
		return nil, err
	}

	diskIDsByName := make(map[string]string, len(disks))
	for _, disk := range disks {
		diskIDsByName[disk.Name] = disk.ID
	}

	mounts := make([]domain.Mount, 0)
	scanner := bufio.NewScanner(strings.NewReader(string(raw)))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		separator := slices.Index(fields, "-")
		if separator == -1 || separator+2 >= len(fields) || len(fields) < 5 {
			continue
		}

		source := fields[separator+2]
		mounts = append(mounts, domain.Mount{
			Source: source,
			Target: fields[4],
			DiskID: diskIDsByName[diskNameFromSource(source)],
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return mounts, nil
}

func diskNameFromSource(source string) string {
	trimmed := strings.TrimPrefix(source, "/dev/")
	trimmed = strings.TrimRight(trimmed, "0123456789")
	return strings.TrimSuffix(trimmed, "p")
}
