package posnode

import (
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

	sync.RWMutex
}

// StartEventEmission starts event emission.
func (n *Node) StartEventEmission() {
	if n.emitter.done != nil {
		return
	}
	n.emitter.done = make(chan struct{})

	n.initParents()

	go func(done chan struct{}) {
		ticker := time.NewTicker(n.conf.EmitInterval)
		for {
			select {
			case <-ticker.C:
				n.EmitEvent()
			case <-done:
				return
			}
		}
	}(n.emitter.done)
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
func (n *Node) AddInternalTxn(tx inter.InternalTransaction) hash.Transaction {
	n.emitter.Lock()
	defer n.emitter.Unlock()

	tx.Index = n.transactionCount() + 1
	tx.Sender = n.ID

	n.emitter.internalTxns = append(n.emitter.internalTxns, &tx)
	return tx.Hash()
}

// transactionCount returns count of internal transactions
// made by the node. If there are no pending transactions
// will ask concensus.
func (n *Node) transactionCount() uint64 {
	txs := n.emitter.internalTxns
	if len(txs) > 0 {
		last := txs[len(txs)-1]
		return last.Index
	}

	return n.consensus.GetTransactionCount(n.ID)
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

	n.Debugf("emiting event")

	var (
		index          uint64
		parents        = hash.Events{}
		maxLamportTime inter.Timestamp
		internalTxns   []*inter.InternalTransaction
		externalTxns   [][]byte
	)

	prev := n.LastEventOf(n.ID)
	if prev != nil {
		index = prev.Index + 1
		maxLamportTime = prev.LamportTime
		parents.Add(prev.Hash())
	} else {
		index = 1
		parents.Add(hash.ZeroEvent)
	}

	for i := 1; i < n.conf.EventParentsCount; i++ {
		p := n.popBestParent()
		if p == nil {
			break
		}
		if !parents.Add(*p) {
			break
		}

		parent := n.store.GetEvent(*p)
		if maxLamportTime < parent.LamportTime {
			maxLamportTime = parent.LamportTime
		}
	}

	// transactions buffer swap
	internalTxns, n.emitter.internalTxns = n.emitter.internalTxns, nil
	externalTxns, n.emitter.externalTxns = n.emitter.externalTxns, nil

	event := &inter.Event{
		Index:                index,
		Creator:              n.ID,
		Parents:              parents,
		LamportTime:          maxLamportTime + 1,
		InternalTransactions: internalTxns,
		ExternalTransactions: externalTxns,
	}
	if err := event.SignBy(n.key); err != nil {
		n.Fatal(err)
	}

	n.saveNewEvent(event, false)
	n.Debugf("new event emited %s", event)

	return event
}
