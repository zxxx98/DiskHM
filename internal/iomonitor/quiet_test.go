package iomonitor_test

import (
	"testing"
	"time"

	"github.com/example/diskhm/internal/iomonitor"
)

func TestQuietEvaluatorRequiresStableWindow(t *testing.T) {
	evaluator := iomonitor.NewQuietEvaluator(10 * time.Second)

	start := time.Unix(100, 0)
	stat := iomonitor.BlockStat{
		ReadsCompleted:  10,
		WritesCompleted: 5,
		InFlight:        0,
	}

	if quiet := evaluator.Advance(start, stat); quiet {
		t.Fatal("expected first sample to be non-quiet")
	}

	if quiet := evaluator.Advance(start.Add(9*time.Second), stat); quiet {
		t.Fatal("expected quiet evaluator to require full stable window")
	}

	if quiet := evaluator.Advance(start.Add(10*time.Second), stat); !quiet {
		t.Fatal("expected quiet evaluator to report quiet after stable window")
	}
}
