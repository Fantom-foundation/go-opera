package originatedtxs

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/hashicorp/golang-lru/simplelru"
)

type Buffer struct {
	senderCount *simplelru.LRU // sender address -> number of transactions
}

func New(maxAddresses int) *Buffer {
	ring := &Buffer{}
	ring.senderCount, _ = simplelru.NewLRU(maxAddresses, nil)
	return ring
}

// Inc is not safe for concurrent use
func (ring *Buffer) Inc(sender common.Address) {
	cur, ok := ring.senderCount.Peek(sender)
	if ok {
		ring.senderCount.Add(sender, cur.(int)+1)
	} else {
		ring.senderCount.Add(sender, int(1))
	}
}

// Dec is not safe for concurrent use
func (ring *Buffer) Dec(sender common.Address) {
	cur, ok := ring.senderCount.Peek(sender)
	if !ok {
		return
	}
	if cur.(int) <= 1 {
		ring.senderCount.Remove(sender)
	} else {
		ring.senderCount.Add(sender, cur.(int)-1)
	}
}

// Clear is not safe for concurrent use
func (ring *Buffer) Clear() {
	ring.senderCount.Purge()
}

// TotalOf is not safe for concurrent use
func (ring *Buffer) TotalOf(sender common.Address) int {
	cur, ok := ring.senderCount.Get(sender)
	if !ok {
		return 0
	}
	return cur.(int)
}

// Empty is not safe for concurrent use
func (ring *Buffer) Empty() bool {
	return ring.senderCount.Len() == 0
}
