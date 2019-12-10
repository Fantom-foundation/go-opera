package inter

import (
	"fmt"
	"math/rand"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// GenNodes generates nodes.
// Result:
//   - nodes  is an array of node addresses;
func GenNodes(
	nodeCount int,
) (
	nodes []idx.StakerID,
) {
	// init results
	nodes = make([]idx.StakerID, nodeCount)
	// make and name nodes
	for i := 0; i < nodeCount; i++ {
		addr := hash.FakePeer()
		nodes[i] = addr
		hash.SetNodeName(addr, "node"+string('A'+i))
	}

	return
}

// ForEachRandFork generates random events with forks for test purpose.
// Result:
//   - callbacks are called for each new event;
//   - events maps node address to array of its events;
func ForEachRandFork(
	nodes []idx.StakerID,
	cheatersArr []idx.StakerID,
	eventCount int,
	parentCount int,
	forksCount int,
	r *rand.Rand,
	callback ForEachEvent,
) (
	events map[idx.StakerID][]*Event,
) {
	if r == nil {
		// fixed seed
		r = rand.New(rand.NewSource(0))
	}
	// init results
	nodeCount := len(nodes)
	events = make(map[idx.StakerID][]*Event, nodeCount)
	cheaters := map[idx.StakerID]int{}
	for _, cheater := range cheatersArr {
		cheaters[cheater] = 0
	}

	// make events
	for i := 0; i < nodeCount*eventCount; i++ {
		// seq parent
		self := i % nodeCount
		creator := nodes[self]
		parents := r.Perm(nodeCount)
		for j, n := range parents {
			if n == self {
				parents = append(parents[0:j], parents[j+1:]...)
				break
			}
		}
		parents = parents[:parentCount-1]
		// make
		e := &Event{
			EventHeader: EventHeader{
				EventHeaderData: EventHeaderData{
					Creator: creator,
					Parents: hash.Events{},
				},
			},
		}
		// first parent is a last creator's event or empty hash
		var parent *Event
		if ee := events[creator]; len(ee) > 0 {
			parent = ee[len(ee)-1]

			// may insert fork
			forksAlready, isCheater := cheaters[creator]
			forkPossible := len(ee) > 1
			forkLimitOk := forksAlready < forksCount
			forkFlipped := r.Intn(eventCount) <= forksCount || i < (nodeCount-1)*eventCount
			if isCheater && forkPossible && forkLimitOk && forkFlipped {
				parent = ee[r.Intn(len(ee)-1)]
				if r.Intn(len(ee)) == 0 {
					parent = nil
				}
				//e.Extra = bigendian.Int32ToBytes(uint32(i)) // make hash for each unique, because for forks we may have the same events
				cheaters[creator]++
			}
		}
		if parent == nil {
			e.Seq = 1
			e.Lamport = 1
		} else {
			e.Seq = parent.Seq + 1
			e.Parents.Add(parent.Hash())
			e.Lamport = parent.Lamport + 1
		}
		// other parents are the lasts other's events
		for _, other := range parents {
			if ee := events[nodes[other]]; len(ee) > 0 {
				parent := ee[len(ee)-1]
				e.Parents.Add(parent.Hash())
				if e.Lamport <= parent.Lamport {
					e.Lamport = parent.Lamport + 1
				}
			}
		}
		name := fmt.Sprintf("%s%03d", string('a'+self), len(events[creator]))
		// buildEvent callback
		if callback.Build != nil {
			e = callback.Build(e, name)
		}
		if e == nil {
			continue
		}
		// calc hash of the event, after it's fully built
		e.RecacheHash()
		e.RecacheSize()
		// save and name event
		hash.SetEventName(e.Hash(), fmt.Sprintf("%s%03d", string('a'+self), len(events[creator])))
		events[creator] = append(events[creator], e)
		// callback
		if callback.Process != nil {
			callback.Process(e, name)
		}
	}

	return
}

// ForEachRandEvent generates random events for test purpose.
// Result:
//   - callbacks are called for each new event;
//   - events maps node address to array of its events;
func ForEachRandEvent(
	nodes []idx.StakerID,
	eventCount int,
	parentCount int,
	r *rand.Rand,
	callback ForEachEvent,
) (
	events map[idx.StakerID][]*Event,
) {
	return ForEachRandFork(nodes, []idx.StakerID{}, eventCount, parentCount, 0, r, callback)
}

// GenRandEvents generates random events for test purpose.
// Result:
//   - events maps node address to array of its events;
func GenRandEvents(
	nodes []idx.StakerID,
	eventCount int,
	parentCount int,
	r *rand.Rand,
) (
	events map[idx.StakerID][]*Event,
) {
	return ForEachRandEvent(nodes, eventCount, parentCount, r, ForEachEvent{})
}

func delPeerIndex(events map[idx.StakerID][]*Event) (res Events) {
	for _, ee := range events {
		res = append(res, ee...)
	}
	return
}
