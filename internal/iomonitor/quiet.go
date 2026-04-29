package iomonitor

import "time"

type QuietEvaluator struct {
	window       time.Duration
	stableSince  time.Time
	lastStat     BlockStat
	hasStableSet bool
}

func NewQuietEvaluator(window time.Duration) *QuietEvaluator {
	return &QuietEvaluator{window: window}
}

func (e *QuietEvaluator) Advance(now time.Time, stat BlockStat) bool {
	if stat.InFlight != 0 {
		e.hasStableSet = false
		e.stableSince = time.Time{}
		e.lastStat = BlockStat{}
		return false
	}

	if !e.hasStableSet || stat.ReadsCompleted != e.lastStat.ReadsCompleted || stat.WritesCompleted != e.lastStat.WritesCompleted {
		e.hasStableSet = true
		e.stableSince = now
		e.lastStat = stat
		return false
	}

	e.lastStat = stat
	return now.Sub(e.stableSince) >= e.window
}
