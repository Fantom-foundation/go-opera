package posnode

import (
	"sync"
	"time"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
)

// emitter creates events from external transactions.
type emitter struct {
	// protects transactions and counter.
	sync.Mutex
	// external transactions.
	transactions [][]byte

	// done allows to stop emitting
	done chan struct{}
}

// StartEmit build events on tick.
func (n *Node) StartEmit() {
	n.emitter.Mutex = sync.Mutex{}
	// n.emitter.transactions = make([]transaction, 0)

	go func() {
		ticker := time.NewTicker(emitterTickInterval)
		for range ticker.C {
			n.CreateEvent()
		}
	}()
}

// StopEmit stops emitter.
func (n *Node) StopEmit() {
	close(n.emitter.done)
}

// AddTransaction adds transaction into the node.
func (n *Node) AddTransaction(t []byte) {
	n.emitter.Lock()
	defer n.emitter.Unlock()

	n.emitter.transactions = append(n.emitter.transactions, t)
}

// CreateEvent takes all transactions from buffer
// builds event, connects it with given amount of
// parents, sign and put it into the storage.
func (n *Node) CreateEvent() {
	n.emitter.Lock()
	defer n.emitter.Unlock()

	if len(n.emitter.transactions) == 0 {
		return
	}

	// Get transactions from buffer and clear it.
	transactions := make([][]byte, len(n.emitter.transactions))
	for i, t := range n.emitter.transactions {
		transactions[i] = []byte(t)
	}
	n.emitter.transactions = make([][]byte, 0)

	lastEvents := n.latestEvents()
	if !enoughEvents(lastEvents, n.conf.EventParentsCount, n.ID.Hex()) {
		n.log.Warn("node does not knew enough events to create new one")
		return
	}

	llt := latestLamportTime(lastEvents)
	parents := selectParents(lastEvents, n.conf.EventParentsCount, n.ID.Hex())

	// Build event.
	event := &inter.Event{
		Index:                n.store.GetPeerHeight(n.ID) + 1,
		Creator:              n.ID,
		Parents:              parents,
		LamportTime:          llt + 1,
		ExternalTransactions: transactions,
	}

	if err := event.SignBy(n.key); err != nil {
		panic(err)
	}

	n.saveNewEvent(event)
}

func (n *Node) latestEvents() []inter.Event {
	heights := n.knownEvents()
	events := make([]inter.Event, 0)
	for creator, index := range heights {
		if index == 0 {
			continue
		}
		eventHash := n.store.GetEventHash(creator, index)
		events = append(events, *n.store.GetEvent(*eventHash))
	}

	return events
}

func enoughEvents(events []inter.Event, count int, parentID string) bool {
	i := 1
	for _, event := range events {
		if event.Creator.Hex() == parentID {
			continue
		}
		i++
	}

	return i >= count
}

func selectParents(events []inter.Event, count int, parentID string) hash.Events {
	var selfEvent *hash.Event
	parents := make([]hash.Event, 0)
	for _, event := range events {
		if event.Creator.Hex() == parentID {
			hash := event.Hash()
			selfEvent = &hash
			continue
		}

		parents = append(parents, event.Hash())
	}

	if selfEvent == nil {
		parents = append(parents, hash.ZeroEvent)
		return hash.NewEvents(parents...)
	}

	parents = append(parents, *selfEvent)
	return hash.NewEvents(parents...)
}

func latestLamportTime(events []inter.Event) inter.Timestamp {
	var lamportTime inter.Timestamp
	for _, event := range events {
		if lamportTime < event.LamportTime {
			lamportTime = event.LamportTime
		}
	}

	return lamportTime
}
