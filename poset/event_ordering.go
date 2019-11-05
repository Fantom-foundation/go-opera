package poset

import (
	"bytes"
	"sort"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

func (p *Poset) frameConsensusTime(frame idx.Frame) inter.Timestamp {
	if frame == 0 {
		return p.PrevEpoch.Time
	}

	return p.store.GetFrameInfo(p.EpochN, frame).
		LastConsensusTime
}

// fareTimestamps calculates time ratio & time offset for the new frame
func (p *Poset) fareTimestamps(
	frame idx.Frame,
	atropos hash.Event,
	highestLamport idx.Lamport,
	lowestLamport idx.Lamport,
) (
	frameInfo FrameInfo,
) {
	lastConsensusTime := p.frameConsensusTime(frame - 1)

	// calculate difference between highest and lowest period
	frameLamportPeriod := idx.MaxLamport(highestLamport-lowestLamport+1, 1)

	// calculate difference between Atropos's median time and previous Atropos's consensus time (almost the same as previous median time)
	nowMedianTime := p.GetEventHeader(p.EpochN, atropos).MedianTime
	frameTimePeriod := inter.MaxTimestamp(nowMedianTime-lastConsensusTime, 1)
	if lastConsensusTime > nowMedianTime {
		frameTimePeriod = 1
	}

	// Calculate time ratio & time offset
	timeRatio := inter.MaxTimestamp(frameTimePeriod/inter.Timestamp(frameLamportPeriod), 1)

	lowestConsensusTime := lastConsensusTime + timeRatio
	timeOffset := int64(lowestConsensusTime) - int64(lowestLamport)*int64(timeRatio)

	// Calculate consensus timestamp of an event with highestLamport (it's always Atropos)
	lastConsensusTime = inter.Timestamp(int64(highestLamport)*int64(timeRatio) + timeOffset)

	// Save new timeRatio & timeOffset to frame
	frameInfo = FrameInfo{
		TimeOffset:        timeOffset,
		TimeRatio:         timeRatio,
		LastConsensusTime: lastConsensusTime,
	}

	return
}

// fareOrdering orders the events
func (p *Poset) fareOrdering(
	unordered []*inter.EventHeaderData,
) (
	ids hash.Events,
) {
	// sort by lamport timestamp & hash
	sort.Slice(unordered, func(i, j int) bool {
		a, b := unordered[i], unordered[j]

		if a.Lamport != b.Lamport {
			return a.Lamport < b.Lamport
		}

		return bytes.Compare(a.Hash().Bytes(), b.Hash().Bytes()) < 0
	})
	ordered := unordered

	ids = make(hash.Events, len(ordered))
	for i, e := range ordered {
		ids[i] = e.Hash()
	}

	return
}
