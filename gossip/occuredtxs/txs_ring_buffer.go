package occuredtxs

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/hashicorp/golang-lru"
)

type buffer struct {
	senders *lru.Cache             // hash -> sender
	from    map[common.Address]int // sender -> number of txs
}

type Buffer struct {
	txSigner types.Signer
	ring     *buffer
}

func New(size int, txSigner types.Signer) *Buffer {
	ring := &buffer{
		from: make(map[common.Address]int),
	}
	ring.senders, _ = lru.NewWithEvict(size, func(key interface{}, value interface{}) {
		ring.decreaseSender(value.(common.Address))
	})
	return &Buffer{
		ring:     ring,
		txSigner: txSigner,
	}
}

// Add is not safe for concurrent use
func (ring *buffer) Add(hash common.Hash, sender common.Address) {
	if ring.senders.Contains(hash) {
		return
	}

	ring.senders.Add(hash, sender)

	ring.from[sender]++
}

// Delete is not safe for concurrent use
func (ring *buffer) Delete(hash common.Hash) {
	ring.senders.Remove(hash)
}

// decreaseSender is not safe for concurrent use
func (ring *buffer) decreaseSender(sender common.Address) {
	was := ring.from[sender]
	if was <= 1 {
		delete(ring.from, sender)
	} else {
		ring.from[sender] = was - 1
	}
}

// Get is not safe for concurrent use
func (ring *buffer) Get(hash common.Hash) (common.Address, bool) {
	sender, ok := ring.senders.Peek(hash)
	if !ok {
		return common.Address{}, false
	}
	return sender.(common.Address), true
}

// GetTxsNum is not safe for concurrent use
func (ring *buffer) GetTxsNum(sender common.Address) int {
	return ring.from[sender]
}

// Clear is not safe for concurrent use
func (ring *buffer) Clear() {
	ring.senders.Purge()
	ring.from = make(map[common.Address]int)
}

// CollectNotConfirmedTxs is called when txs are included into an event, but not included into a block
// not safe for concurrent use
func (s *Buffer) CollectNotConfirmedTxs(txs types.Transactions) error {
	for _, tx := range txs {
		sender, err := s.txSigner.Sender(tx)
		if err != nil {
			return err
		}
		s.ring.Add(tx.Hash(), sender)
	}
	return nil
}

// CollectConfirmedTxs is called when txs are included into a block
// not safe for concurrent use
func (s *Buffer) CollectConfirmedTxs(txs types.Transactions) {
	for _, tx := range txs {
		s.ring.Delete(tx.Hash())
	}
}

// MayBeConflicted is not safe for concurrent use
func (s *Buffer) MayBeConflicted(sender common.Address, txHash common.Hash) bool {
	_, ok := s.ring.Get(txHash)
	if ok {
		return true // the tx was already included by somebody
	}
	// the sender already has a not confirmed tx, wait until it gets confirmed
	return s.ring.GetTxsNum(sender) != 0
}

// Clear is not safe for concurrent use
func (s *Buffer) Clear() {
	s.ring.Clear()
}

// Len is safe for concurrent use
func (s *Buffer) Len() int {
	return s.ring.senders.Len()
}
