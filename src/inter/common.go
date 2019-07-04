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
func GenEventsByNode(
	nodeCount int,
	eventCount int,
	parentCount int,
	mods ...func(*Event, []hash.Peer),
) (
	nodes []hash.Peer,
	events map[hash.Peer][]*Event,
) {
	// init results
	nodes = make([]hash.Peer, nodeCount)
	events = make(map[hash.Peer][]*Event, nodeCount)
	// make and name nodes
	for i := 0; i < nodeCount; i++ {
		addr := hash.FakePeer()
		nodes[i] = addr
		hash.NodeNameDict[addr] = "node" + string('A'+i)
	}
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
			Creator: creator,
			Parents: hash.Events{},
		}
		// first parent is a last creator's event or empty hash
		if ee := events[creator]; len(ee) > 0 {
			parent := ee[len(ee)-1]
			e.Index = parent.Index + 1
			e.Parents.Add(parent.Hash())
			e.LamportTime = parent.LamportTime + 1
		} else {
			e.Index = 1
			e.Parents.Add(hash.ZeroEvent)
			e.LamportTime = 1
		}
		// other parents are the lasts other's events
		for _, other := range parents {
			if ee := events[nodes[other]]; len(ee) > 0 {
				parent := ee[len(ee)-1]
				e.Parents.Add(parent.Hash())
				if e.LamportTime <= parent.LamportTime {
					e.LamportTime = parent.LamportTime + 1
				}
			}
		}
		// apply mods
		for _, mod := range mods {
			mod(e, nodes)
		}
		// save and name event
		hash.EventNameDict[e.Hash()] = fmt.Sprintf("%s%03d", string('a'+self), len(events[creator]))
		events[creator] = append(events[creator], e)
	}

	return
}

func delPeerIndex(events map[hash.Peer][]*Event) (res Events) {
	for _, ee := range events {
		res = append(res, ee...)
	}
	return
}
