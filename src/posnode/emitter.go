package posnode

import (
	"fmt"
	"sync"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// emitter creates events from external transactions.
type emitter struct {
	internalTxns map[hash.Transaction]*inter.InternalTransaction
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

func (n *Node) internalTxnMempool(idx hash.Transaction) *inter.InternalTransaction {
	n.emitter.RLock()
	defer n.emitter.RUnlock()

	// chck mempool
	if n.emitter.internalTxns == nil {
		return nil
	}
	return n.emitter.internalTxns[idx]
}

// AddInternalTxn takes internal transaction for new event.
func (n *Node) AddInternalTxn(tx inter.InternalTransaction) (hash.Transaction, error) {
	if tx.Receiver == n.ID {
		return hash.Transaction{}, fmt.Errorf("can not transfer to yourself")
	}

	if tx.Amount < 1 {
		return hash.Transaction{}, fmt.Errorf("can not transfer zero amount")
	}

	if balance := n.consensus.StakeOf(n.ID); tx.Amount > balance {
		return hash.Transaction{}, fmt.Errorf("insufficient funds %d to transfer %d", balance, tx.Amount)
	}

	idx := inter.TransactionHashOf(n.ID, tx.Index)

	n.emitter.Lock()
	defer n.emitter.Unlock()

	if n.emitter.internalTxns == nil {
		n.emitter.internalTxns = make(map[hash.Transaction]*inter.InternalTransaction)
	}

	if prev, ok := n.emitter.internalTxns[idx]; ok {
		return idx, fmt.Errorf("the same txn is in mempool already: %+v", prev)
	}

	if e := n.store.GetTxnsEvent(idx); e != nil {
		return idx, fmt.Errorf("the same txn already exists in event %d of %s", e.Index, e.Creator.String())
	}

	n.emitter.internalTxns[idx] = &tx

	return idx, nil
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
// It returns emitted event for test purpose.
func (n *Node) EmitEvent() *inter.Event {
	n.emitter.Lock()
	defer n.emitter.Unlock()

	n.Debugf("emitting event")

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
	internalTxns = make([]*inter.InternalTransaction, 0, len(n.emitter.internalTxns))
	for idx, txn := range n.emitter.internalTxns {
		n.Debugf("event internal tx [%s] amount: %d from [%s] to [%s]",
			idx.Hex(), txn.Amount, n.ID.Hex(), txn.Receiver.Hex())
		internalTxns = append(internalTxns, txn)
	}
	n.emitter.internalTxns = nil

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

	n.onNewEvent(event)
	n.Infof("new event emitted %s", event)

	return event
}
