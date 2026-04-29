package discovery

import (
	"io/fs"
	"strings"
)

type udevInfo struct {
	Model     string
	Serial    string
	Transport string
}

func readUdevInfo(fsys fs.FS, name string) (udevInfo, error) {
	raw, err := fs.ReadFile(fsys, "run/udev/data/b8:0")
	if err != nil {
		fallback, fallbackErr := fs.ReadFile(fsys, "run/udev/data/b8_0")
		if fallbackErr != nil {
			return udevInfo{}, err
		}
		raw = fallback
	}

	var info udevInfo
	lines := strings.Split(string(raw), "\n")
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "E:ID_MODEL="):
			info.Model = parseUdevValue(strings.TrimPrefix(line, "E:ID_MODEL="))
		case strings.HasPrefix(line, "E:ID_SERIAL_SHORT="):
			info.Serial = parseUdevValue(strings.TrimPrefix(line, "E:ID_SERIAL_SHORT="))
		case strings.HasPrefix(line, "E:ID_BUS="):
			info.Transport = parseUdevValue(strings.TrimPrefix(line, "E:ID_BUS="))
		}
	}

	_ = name
	return info, nil
}

func parseUdevValue(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), "_", " ")
}
