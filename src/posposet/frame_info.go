package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// TODO: make FrameInfo internal

type FrameInfo struct {
	TimeOffset inter.Timestamp
	TimeRatio  inter.Timestamp
}

// GetConsensusTimestamp calc consensus timestamp for given event.
func (f *FrameInfo) GetConsensusTimestamp(e *Event) inter.Timestamp {
	return inter.Timestamp(e.Lamport)*f.TimeOffset + f.TimeRatio
}
