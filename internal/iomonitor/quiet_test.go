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

func TestQuietEvaluatorInFlightClearsStability(t *testing.T) {
	evaluator := iomonitor.NewQuietEvaluator(10 * time.Second)

	start := time.Unix(100, 0)
	stableStat := iomonitor.BlockStat{
		ReadsCompleted:  10,
		WritesCompleted: 5,
	}

	evaluator.Advance(start, stableStat)
	if quiet := evaluator.Advance(start.Add(10*time.Second), stableStat); !quiet {
		t.Fatal("expected quiet after initial stable window")
	}

	inFlightStat := stableStat
	inFlightStat.InFlight = 1
	if quiet := evaluator.Advance(start.Add(11*time.Second), inFlightStat); quiet {
		t.Fatal("expected in-flight sample to clear quiet state")
	}

	if quiet := evaluator.Advance(start.Add(20*time.Second), stableStat); quiet {
		t.Fatal("expected fresh stable sample after in-flight activity to restart quiet window")
	}

	if quiet := evaluator.Advance(start.Add(30*time.Second), stableStat); !quiet {
		t.Fatal("expected quiet only after a full new stable window")
	}
}

func TestQuietEvaluatorReadChangeRestartsWindow(t *testing.T) {
	evaluator := iomonitor.NewQuietEvaluator(10 * time.Second)

	start := time.Unix(100, 0)
	stat := iomonitor.BlockStat{
		ReadsCompleted:  10,
		WritesCompleted: 5,
	}

	evaluator.Advance(start, stat)
	if quiet := evaluator.Advance(start.Add(10*time.Second), stat); !quiet {
		t.Fatal("expected quiet after initial stable window")
	}

	changed := stat
	changed.ReadsCompleted++
	if quiet := evaluator.Advance(start.Add(11*time.Second), changed); quiet {
		t.Fatal("expected read counter change to restart quiet window")
	}

	if quiet := evaluator.Advance(start.Add(20*time.Second), changed); quiet {
		t.Fatal("expected read-stable period shorter than window to remain non-quiet")
	}

	if quiet := evaluator.Advance(start.Add(21*time.Second), changed); !quiet {
		t.Fatal("expected quiet after read counter remains stable for a full new window")
	}
}

func TestQuietEvaluatorWriteChangeRestartsWindow(t *testing.T) {
	evaluator := iomonitor.NewQuietEvaluator(10 * time.Second)

	start := time.Unix(100, 0)
	stat := iomonitor.BlockStat{
		ReadsCompleted:  10,
		WritesCompleted: 5,
	}

	evaluator.Advance(start, stat)
	if quiet := evaluator.Advance(start.Add(10*time.Second), stat); !quiet {
		t.Fatal("expected quiet after initial stable window")
	}

	changed := stat
	changed.WritesCompleted++
	if quiet := evaluator.Advance(start.Add(11*time.Second), changed); quiet {
		t.Fatal("expected write counter change to restart quiet window")
	}

	if quiet := evaluator.Advance(start.Add(20*time.Second), changed); quiet {
		t.Fatal("expected write-stable period shorter than window to remain non-quiet")
	}

	if quiet := evaluator.Advance(start.Add(21*time.Second), changed); !quiet {
		t.Fatal("expected quiet after write counter remains stable for a full new window")
	}
}
