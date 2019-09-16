package poset

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// TODO: make FrameInfo internal

type FrameInfo struct {
	TimeOffset int64 // may be negative
	TimeRatio  inter.Timestamp
}

// GetConsensusTimestamp calc consensus timestamp for given event.
func (f *FrameInfo) GetConsensusTimestamp(e *Event) inter.Timestamp {
	return inter.Timestamp(int64(e.Lamport)*int64(f.TimeRatio) + f.TimeOffset)
}
