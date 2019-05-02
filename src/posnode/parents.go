package posnode

import (
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

type (
	parent struct {
		Creator hash.Peer
		Parents hash.Events
		Value   float64
		Last    bool
	}

	// parents is a potential parent events cache.
	parents struct {
		cache map[hash.Event]*parent
		sync.Mutex
	}
)

func (n *Node) initParents() {
	const loadDeep uint64 = 10

	n.parents.Lock()
	defer n.parents.Unlock()

	if n.parents.cache != nil {
		return
	}

	n.parents.cache = make(map[hash.Event]*parent)

	// load some parents from store
	for _, peer := range n.peers.Snapshot() {
		to := n.store.GetPeerHeight(peer)
		from := uint64(1)
		if (from + loadDeep) <= to {
			from -= loadDeep
		}
		for i := from; i <= to; i++ {
			e := n.EventOf(peer, i)
			val := float64(1)
			if n.consensus != nil {
				val = n.consensus.GetStakeOf(e.Creator)
			}
			n.parents.cache[e.Hash()] = &parent{
				Creator: e.Creator,
				Parents: e.Parents,
				Value:   val,
				Last:    (i == to),
			}
		}
	}

}

// pushPotentialParent add event to parent events cache.
// TODO: check events order, their parents should be pushed first ( see posposet/Poset.onNewEvent() ).
func (n *Node) pushPotentialParent(e *inter.Event) {
	n.parents.Lock()
	defer n.parents.Unlock()

	if n.parents.cache == nil {
		return
	}

	if e.Creator == n.ID {
		return
	}

	if _, ok := n.parents.cache[e.Hash()]; ok {
		return
	}

	val := float64(1)
	if n.consensus != nil {
		val = n.consensus.GetStakeOf(e.Creator)
	}

	n.parents.cache[e.Hash()] = &parent{
		Creator: e.Creator,
		Parents: e.Parents,
		Value:   val,
		Last:    true,
	}

	prev := n.store.GetEventHash(e.Creator, e.Index-1)
	if prev != nil {
		n.parents.cache[*prev].Last = false
	}
}

// popBestParent returns best parent and marks it as used.
func (n *Node) popBestParent() *hash.Event {
	n.parents.Lock()
	defer n.parents.Unlock()

	var (
		res *hash.Event
		max float64
		tmp hash.Event
	)

	for e, p := range n.parents.cache {
		if !p.Last {
			continue
		}

		val := n.parents.Sum(e)
		if val > max {
			tmp, res = e, &tmp
			max = val
		}
	}

	if res != nil {
		n.parents.Del(*res)
	}

	return res
}

/*
 * parents utils:
 */

// Sum returns sum of parent values.
func (pp *parents) Sum(e hash.Event) float64 {
	event, ok := pp.cache[e]
	if !ok {
		return float64(0)
	}

	res := event.Value
	for p := range event.Parents {
		res += pp.Sum(p)
	}

	return res
}

// Del removes whole event tree.
func (pp *parents) Del(e hash.Event) {
	event, ok := pp.cache[e]
	if !ok {
		return
	}

	for p := range event.Parents {
		pp.Del(p)
	}

	delete(pp.cache, e)
}
