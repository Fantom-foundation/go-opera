package posnode

import (
	"sort"
	"sync"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// emitter creates events from external transactions.
type emitter struct {
	internalTxns []*inter.InternalTransaction
	externalTxns [][]byte
	done         chan struct{}

	sync.Mutex
}

// StartEventEmission starts event emission.
func (n *Node) StartEventEmission() {
	if n.emitter.done != nil {
		return
	}
	n.emitter.done = make(chan struct{})
	done := n.emitter.done

	go func() {
		ticker := time.NewTicker(n.conf.EmitInterval)
		for {
			select {
			case <-ticker.C:
				n.EmitEvent()
			case <-done:
				return
			}
		}
	}()
}

// StopEventEmission stops event emission.
func (n *Node) StopEventEmission() {
	if n.emitter.done == nil {
		return
	}

	close(n.emitter.done)
	n.emitter.done = nil
}

// AddInternalTxn takes internal transaction for new event.
func (n *Node) AddInternalTxn(tx inter.InternalTransaction) {
	n.emitter.Lock()
	defer n.emitter.Unlock()

	n.emitter.internalTxns = append(n.emitter.internalTxns, &tx)
}

// AddExternalTxn takes external transaction for new event.
func (n *Node) AddExternalTxn(tx []byte) {
	n.emitter.Lock()
	defer n.emitter.Unlock()
	// TODO: copy tx val?
	n.emitter.externalTxns = append(n.emitter.externalTxns, tx)
}

// EmitEvent takes all transactions from buffer builds event,
// connects it with given amount of parents, sign and put it into the storage.
// It returns emmited event for test purpose.
func (n *Node) EmitEvent() *inter.Event {
	n.emitter.Lock()
	defer n.emitter.Unlock()

	var (
		index        uint64
		parents      hash.Events = hash.Events{}
		lamportTime  inter.Timestamp
		internalTxns []*inter.InternalTransaction
		externalTxns [][]byte
	)

	// transactions buffer swap
	internalTxns, n.emitter.internalTxns = n.emitter.internalTxns, nil
	externalTxns, n.emitter.externalTxns = n.emitter.externalTxns, nil

	// ref nodes selection
	refs := n.peers.Snapshot()
	sort.Sort(n.emitterEvaluation(refs))
	count := n.conf.EventParentsCount - 1
	if len(refs) > count {
		refs = refs[:count]
	}
	refs = append(refs, n.ID)

	// last events of ref nodes
	for _, ref := range refs {
		h := n.store.GetPeerHeight(ref)
		if h < 1 {
			if ref == n.ID {
				index = 1
				parents.Add(hash.ZeroEvent)
			}
			continue
		}

		if ref == n.ID {
			index = h + 1
		}

		e := n.store.GetEventHash(ref, h)
		if e == nil {
			n.log.Errorf("no event hash for (%s,%d) in store", ref.String(), h)
			continue
		}
		event := n.store.GetEvent(*e)
		if event == nil {
			n.log.Errorf("no event %s in store", e.String())
			continue
		}

		parents.Add(*e)
		n.UsedAsParent(ref)
		if lamportTime < event.LamportTime {
			lamportTime = event.LamportTime
		}
	}

	event := &inter.Event{
		Index:                index,
		Creator:              n.ID,
		Parents:              parents,
		LamportTime:          lamportTime + 1,
		InternalTransactions: internalTxns,
		ExternalTransactions: externalTxns,
	}
	if err := event.SignBy(n.key); err != nil {
		panic(err)
	}

	n.saveNewEvent(event)

	return event
}

/*
 * evaluation function for emitter
 */

func (n *Node) emitterEvaluation(peers []hash.Peer) *emitterEvaluation {
	return &emitterEvaluation{
		node:  n,
		peers: peers,
	}
}

// emitterEvaluation implements sort.Interface.
type emitterEvaluation struct {
	node  *Node
	peers []hash.Peer
}

// Len is the number of elements in the collection.
func (e *emitterEvaluation) Len() int {
	return len(e.peers)
}

// Swap swaps the elements with indexes i and j.
func (e *emitterEvaluation) Swap(i, j int) {
	e.peers[i], e.peers[j] = e.peers[j], e.peers[i]
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (e *emitterEvaluation) Less(i, j int) bool {
	a := e.node.peers.attrByID(e.peers[i])
	b := e.node.peers.attrByID(e.peers[j])

	return a.LastUsed.Before(b.LastUsed) ||
		a.LastEvent.After(b.LastEvent) ||
		e.node.consensus.GetStakeOf(e.peers[i]) < e.node.consensus.GetStakeOf(e.peers[j])
}
