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

			if a.LamportTime != b.LamportTime {
				return a.LamportTime < b.LamportTime
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

		if event.LamportTime > latestEvents[event.Creator].LamportTime {
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

	// 3. Get Stake from each creator
	stakes := map[hash.Peer]inter.Stake{}
	var jointStake inter.Stake
	for _, event := range selectedEvents {
		stake := p.StakeOf(event.Creator)

		stakes[event.Creator] = stake
		jointStake += stake
	}

	halfStake := jointStake / 2

	// 4. Calculate weighted median
	selectedEventsMap := map[hash.Peer]*inter.Event{}
	for _, event := range selectedEvents {
		selectedEventsMap[event.Creator] = event
	}

	var currStake inter.Stake
	var median *inter.Event
	for node, stake := range stakes {
		if currStake < halfStake {
			currStake += stake
			continue
		}

		median = selectedEventsMap[node]
		break
	}

	highestTimestamp := selectedEvents[len(selectedEvents)-1].LamportTime
	lowestTimestamp := selectedEvents[0].LamportTime

	// 5. Calculate time ratio & time offset
	if p.LastConsensusTime == 0 {
		p.LastConsensusTime = genesisTimestamp
	}

	frameTimePeriod := inter.MaxTimestamp(median.LamportTime-p.LastConsensusTime, 1)
	frameLamportPeriod := inter.MaxTimestamp(highestTimestamp-lowestTimestamp, 1)

	timeRatio := inter.MaxTimestamp(frameTimePeriod/frameLamportPeriod, 1)

	lowestConsensusTime := p.LastConsensusTime + timeRatio
	timeOffset := lowestConsensusTime - lowestTimestamp*timeRatio

	// 6. Calculate consensus timestamp
	p.LastConsensusTime = p.input.GetEvent(sfWitness).LamportTime*timeRatio + timeOffset

	// 7. Save new timeRatio & timeOffset to frame
	p.frames[frame].SetTimes(timeOffset, timeRatio)

	return sortEvents(unordered)
}
