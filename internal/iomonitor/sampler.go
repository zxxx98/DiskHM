package iomonitor

type BlockStat struct {
	ReadsCompleted  uint64
	WritesCompleted uint64
	InFlight        uint64
}
