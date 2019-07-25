package posnode

import (
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

type (
	parent struct {
		Creator hash.Peer
		Parents hash.Events
		Value   inter.Stake
		Last    bool
	}

	// parents is a potential parent events cache.
	parents struct {
		cache map[hash.Event]*parent
		sync.RWMutex
	}
)

func (n *Node) initParents() {
	n.initSuperFrame()

	const loadDeep idx.Event = 10

	n.parents.Lock()
	defer n.parents.Unlock()

	if n.parents.cache != nil {
		return
	}
	n.parents.cache = make(map[hash.Event]*parent)

	sf := n.currentSuperFrame()

	// load some parents from store
	for _, peer := range n.peers.Snapshot() {
		to := n.store.GetPeerHeight(peer, sf)
		from := idx.Event(1)
		if (from + loadDeep) <= to {
			from -= loadDeep
		}
		for i := from; i <= to; i++ {
			e := n.EventOf(peer, sf, i)
			val := inter.Stake(1)
			if n.consensus != nil {
				val = n.consensus.StakeOf(e.Creator)
			}
			n.parents.cache[e.Hash()] = &parent{
				Creator: e.Creator,
				Parents: e.Parents,
				Value:   val,
				Last:    i == to,
			}
		}
	}

}

// pushPotentialParent adds event to parent events cache except self-events.
// Parents should be pushed first ( see posposet/Poset.onNewEvent() ).
func (n *Node) pushPotentialParent(e *inter.Event) {
	if e.Creator == n.ID {
		return
	}

	val := inter.Stake(1)
	if n.consensus != nil {
		val = n.consensus.StakeOf(e.Creator)
	}

	n.parents.Push(e, val)
}

// Push adds parent to cache.
func (pp *parents) Push(e *inter.Event, val inter.Stake) {
	pp.Lock()
	defer pp.Unlock()

	if pp.cache == nil {
		return
	}

	if _, ok := pp.cache[e.Hash()]; ok {
		return
	}

	for p := range e.Parents {
		if prev, ok := pp.cache[p]; ok {
			prev.Last = false
		}
	}

	pp.cache[e.Hash()] = &parent{
		Creator: e.Creator,
		Parents: e.Parents,
		Value:   val,
		Last:    true,
	}
}

// PopBest returns best parent and marks it as used.
func (pp *parents) PopBest() *hash.Event {
	pp.Lock()
	defer pp.Unlock()

	var (
		res *hash.Event
		max inter.Stake
		tmp hash.Event
	)

	for e, p := range pp.cache {
		if !p.Last {
			continue
		}

		val := pp.sum(e)
		if val > max {
			tmp, res = e, &tmp
			max = val
		}
	}

	if res != nil {
		pp.del(*res)
	}

	return res
}

// Count of potential parents.
func (pp *parents) Count() int {
	pp.Lock()
	defer pp.Unlock()

	if pp.cache == nil {
		return 0
	}

	n := 0
	for _, p := range pp.cache {
		if p.Last {
			n++
		}
	}
	return n
}

/*
 * parents utils:
 */

// sum returns sum of parent values.
func (pp *parents) sum(e hash.Event) inter.Stake {
	event, ok := pp.cache[e]
	if !ok {
		return inter.Stake(0)
	}

	res := event.Value
	for p := range event.Parents {
		res += pp.sum(p)
	}

	return res
}

// del removes whole event tree.
func (pp *parents) del(e hash.Event) {
	event, ok := pp.cache[e]
	if !ok {
		return
	}

	for p := range event.Parents {
		pp.del(p)
	}

	delete(pp.cache, e)
}
