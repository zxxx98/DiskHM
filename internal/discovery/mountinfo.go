package discovery

import (
	"bufio"
	"io/fs"
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
		if len(fields) < 10 {
			continue
		}

		source := fields[8]
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
	return strings.TrimRight(trimmed, "0123456789")
}
