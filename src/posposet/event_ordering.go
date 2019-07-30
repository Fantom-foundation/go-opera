package posposet

import (
	"bytes"
	"sort"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
)

const (
	nodeCount                        = internal.MembersCount / 3
	genesisTimestamp inter.Timestamp = 1562816974
)

func (p *Poset) fareOrdering(frame idx.Frame, sfWitness hash.Event, unordered inter.Events) inter.Events {
	// sort by lamport timestamp & hash
	sortEvents := func(events []*inter.Event) []*inter.Event {
		sort.Slice(events, func(i, j int) bool {
			a, b := events[i], events[j]

			if a.Lamport != b.Lamport {
				return a.Lamport < b.Lamport
			}

			return bytes.Compare(a.Hash().Bytes(), b.Hash().Bytes()) < 0
		})

		return events
	}

	// 1. Select latest events from each node with greatest lamport timestamp
	latestEvents := map[hash.Peer]*inter.Event{}
	for _, event := range unordered {
		if _, ok := latestEvents[event.Creator]; !ok {
			latestEvents[event.Creator] = event
			continue
		}

		if event.Lamport > latestEvents[event.Creator].Lamport {
			latestEvents[event.Creator] = event
		}
	}

	// 2. Sort by lamport
	var selectedEvents []*inter.Event
	for _, event := range latestEvents {
		selectedEvents = append(selectedEvents, event)
	}

	selectedEvents = sortEvents(selectedEvents)

	if len(selectedEvents) > nodeCount {
		selectedEvents = selectedEvents[:nodeCount-1]
	}

	highestTimestamp := selectedEvents[len(selectedEvents)-1].ClaimedTime
	lowestTimestamp := selectedEvents[0].ClaimedTime

	// 3. Calculate time ratio & time offset
	if p.LastConsensusTime == 0 {
		p.LastConsensusTime = genesisTimestamp
	}

	frameTimePeriod := inter.MaxTimestamp(p.vi.MedianTime(sfWitness, genesisTimestamp)-p.LastConsensusTime, 1)
	frameLamportPeriod := inter.MaxTimestamp(highestTimestamp-lowestTimestamp, 1)

	timeRatio := inter.MaxTimestamp(frameTimePeriod/frameLamportPeriod, 1)

	lowestConsensusTime := p.LastConsensusTime + timeRatio
	timeOffset := lowestConsensusTime - lowestTimestamp*timeRatio

	// 4. Calculate consensus timestamp
	p.LastConsensusTime = inter.Timestamp(p.input.GetEvent(sfWitness).Lamport)*timeRatio + timeOffset

	// 5. Save new timeRatio & timeOffset to frame
	p.frames[frame].SetTimes(timeOffset, timeRatio)

	return sortEvents(unordered)
}
