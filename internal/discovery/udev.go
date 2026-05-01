package discovery

import (
	"io/fs"
	"path/filepath"
	"strings"
)

type udevInfo struct {
	Model     string
	Serial    string
	Transport string
}

func readUdevInfo(fsys fs.FS, name string) (udevInfo, error) {
	deviceNumber, err := readDeviceNumber(fsys, name)
	if err != nil {
		return udevInfo{}, err
	}

	primaryPath := filepath.ToSlash(filepath.Join("run/udev/data", "b"+deviceNumber))
	raw, err := fs.ReadFile(fsys, primaryPath)
	if err != nil {
		fallbackPath := filepath.ToSlash(filepath.Join("run/udev/data", "b"+strings.ReplaceAll(deviceNumber, ":", "_")))
		fallback, fallbackErr := fs.ReadFile(fsys, fallbackPath)
		if fallbackErr != nil {
			return udevInfo{}, nil
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
	return info, nil
}

func parseUdevValue(value string) string {
	return strings.ReplaceAll(strings.TrimSpace(value), "_", " ")
}
