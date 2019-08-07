package inter

import (
	"fmt"
	"math/rand"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

type (
	// Stake amount.
	Stake uint64
)

// GenEventsByNode generates random events for test purpose.
// Result:
//   - nodes  is an array of node addresses;
//   - events maps node address to array of its events;
func GenNodes(
	nodeCount int,
) (
	nodes []hash.Peer,
) {
	// init results
	nodes = make([]hash.Peer, nodeCount)
	// make and name nodes
	for i := 0; i < nodeCount; i++ {
		addr := hash.FakePeer()
		nodes[i] = addr
		hash.SetNodeName(addr, "node"+string('A'+i))
	}

	return
}

// GenEventsByNode generates random events for test purpose.
// Result:
//   - nodes  is an array of node addresses;
//   - events maps node address to array of its events;
func GenEventsByNode(
	nodes []hash.Peer,
	eventCount int,
	parentCount int,
	buildEvent func(*Event) *Event,
	onNewEvent func(*Event),
) (
	events map[hash.Peer][]*Event,
) {
	// init results
	nodeCount := len(nodes)
	events = make(map[hash.Peer][]*Event, nodeCount)
	// make events
	for i := 0; i < nodeCount*eventCount; i++ {
		// seq parent
		self := i % nodeCount
		creator := nodes[self]
		parents := rand.Perm(nodeCount)
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
		if ee := events[creator]; len(ee) > 0 {
			parent := ee[len(ee)-1]
			e.Seq = parent.Seq + 1
			e.Parents.Add(parent.Hash())
			e.Lamport = parent.Lamport + 1
		} else {
			e.Seq = 1
			e.Lamport = 1
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
		// buildEvent callback
		if buildEvent != nil {
			e = buildEvent(e)
		}
		if e == nil {
			continue
		}
		// calc hash of the event, after it's fully built
		e.RecacheHash()
		// save and name event
		hash.SetEventName(e.Hash(), fmt.Sprintf("%s%03d", string('a'+self), len(events[creator])))
		events[creator] = append(events[creator], e)
		// callback
		if onNewEvent != nil {
			onNewEvent(e)
		}
	}

	return
}

func delPeerIndex(events map[hash.Peer][]*Event) (res Events) {
	for _, ee := range events {
		res = append(res, ee...)
	}
	return
}
