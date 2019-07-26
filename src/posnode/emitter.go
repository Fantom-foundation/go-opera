package posnode

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
)

// emitter creates events from external transactions.
type emitter struct {
	internalTxns map[hash.Transaction]*inter.InternalTransaction
	externalTxns [][]byte

	last time.Time
	sync.RWMutex
	done chan struct{}
	wg   sync.WaitGroup
}

// StartEventEmission starts event emission.
func (n *Node) StartEventEmission() {
	if n.emitter.done != nil {
		return
	}
	n.emitter.done = make(chan struct{})

	n.initParents()

	done := n.emitter.done
	n.emitter.wg.Add(1)
	go func() {
		defer n.emitter.wg.Done()
		ticker := time.NewTicker(n.conf.MinEmitInterval)
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
	n.emitter.wg.Wait()
}

func (n *Node) internalTxnMempool(idx hash.Transaction) *inter.InternalTransaction {
	n.emitter.RLock()
	defer n.emitter.RUnlock()

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

	idx := inter.TransactionHashOf(n.ID, tx.Nonce)

	n.emitter.Lock()
	defer n.emitter.Unlock()

	if n.emitter.internalTxns == nil {
		n.emitter.internalTxns = make(map[hash.Transaction]*inter.InternalTransaction)
	}

	if prev, ok := n.emitter.internalTxns[idx]; ok {
		return idx, fmt.Errorf("the same txn is in mempool already: %+v", prev)
	}

	if e := n.store.GetTxnsEvent(idx); e != nil {
		return idx, fmt.Errorf("the same txn already exists in event %s of %s", e.Hash().String(), e.Creator.String())
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

	if time.Now().Add(-n.conf.MaxEmitInterval).Before(n.emitter.last) &&
		n.parents.Count() < (n.conf.EventParentsCount-1) &&
		len(n.emitter.externalTxns) < 1 &&
		len(n.emitter.internalTxns) < 1 {
		n.Debugf("nothing to emit")
		return nil
	}

	return n.emitEvent()
}

// emitEvent with no checks.
func (n *Node) emitEvent() *inter.Event {
	var (
		sf             = n.currentSuperFrame()
		seq            idx.Event
		selfParent     hash.Event
		parents        = hash.Events{}
		maxLamportTime inter.Timestamp
		internalTxns   []*inter.InternalTransaction
		externalTxns   [][]byte
	)

	prev := n.LastEventOf(n.ID, sf)
	if prev != nil {
		seq = prev.Seq + 1
		maxLamportTime = prev.LamportTime
		selfParent = prev.Hash()
		parents.Add(prev.Hash())
	} else {
		seq = 1
		selfParent = hash.ZeroEvent
		parents.Add(hash.ZeroEvent)
	}

	for i := 1; i < n.conf.EventParentsCount; i++ {
		p := n.parents.PopBest()
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

	// TODO: don't use txns if event has no chance to be in block
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
		SfNum:                sf,
		Seq:                  seq,
		Creator:              n.ID,
		SelfParent:           selfParent,
		Parents:              parents,
		LamportTime:          maxLamportTime + 1,
		InternalTransactions: internalTxns,
		ExternalTransactions: inter.ExtTxns{
			Value: externalTxns,
		},
	}
	if err := event.SignBy(n.key); err != nil {
		n.Fatal(err)
	}

	// set event name for debug
	n.nameEventForDebug(event)

	n.emitter.last = time.Now()
	countEmittedEvents.Inc(1)

	n.onNewEvent(event)
	n.Infof("new event emitted %s", event)

	return event
}

func (n *Node) nameEventForDebug(e *inter.Event) {
	name := []rune(hash.GetNodeName(n.ID))
	if len(name) < 1 {
		return
	}

	name = name[len(name)-1:]
	hash.SetEventName(e.Hash(), fmt.Sprintf("%s%03d",
		strings.ToLower(string(name)),
		e.Seq))
}
