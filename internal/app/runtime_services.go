package app

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/example/diskhm/internal/domain"
	"github.com/example/diskhm/internal/iomonitor"
)

type noopProbe struct{}

func (noopProbe) Safe(context.Context, domain.Disk) error {
	return nil
}

func (noopProbe) Wake(context.Context, domain.Disk) error {
	return nil
}

type syncFlusher struct{}

func (syncFlusher) Flush(ctx context.Context, disk domain.Disk) error {
	_ = disk
	return exec.CommandContext(ctx, "sync").Run()
}

type sysfsQuietSampler struct {
	root   string
	window time.Duration
}

func (s sysfsQuietSampler) IsQuiet(ctx context.Context, disk domain.Disk) (bool, error) {
	window := s.window
	if window <= 0 {
		window = time.Second
	}

	evaluator := iomonitor.NewQuietEvaluator(window)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		stat, err := readBlockStat(filepath.Join(s.root, disk.Name, "stat"))
		if err != nil {
			return false, err
		}
		if evaluator.Advance(time.Now(), stat) {
			return true, nil
		}

		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-ticker.C:
		}
	}
}

func readBlockStat(path string) (iomonitor.BlockStat, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return iomonitor.BlockStat{}, err
	}

	fields := strings.Fields(string(data))
	if len(fields) < 9 {
		return iomonitor.BlockStat{}, errors.New("unexpected /sys/block stat format")
	}

	readsCompleted, err := strconv.ParseUint(fields[0], 10, 64)
	if err != nil {
		return iomonitor.BlockStat{}, err
	}
	writesCompleted, err := strconv.ParseUint(fields[4], 10, 64)
	if err != nil {
		return iomonitor.BlockStat{}, err
	}
	inFlight, err := strconv.ParseUint(fields[8], 10, 64)
	if err != nil {
		return iomonitor.BlockStat{}, err
	}

	return iomonitor.BlockStat{
		ReadsCompleted:  readsCompleted,
		WritesCompleted: writesCompleted,
		InFlight:        inFlight,
	}, nil
}
