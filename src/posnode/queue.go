package posnode

import (
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
)

type queue struct {
	heights map[string]bool

	sync sync.Mutex
}

type parentQueue struct {
	hashes map[hash.Event]bool

	sync sync.Mutex
}

func initQueue() queue {
	return queue{
		heights: map[string]bool{},
	}
}

func initParentQueue() parentQueue {
	return parentQueue{
		hashes: map[hash.Event]bool{},
	}
}

// Main Queue

// AddToQueue known peer height
func (n *Node) addToQueue(key string) {
	n.queue.sync.Lock()
	defer n.queue.sync.Unlock()

	n.queue.heights[key] = true
}

// DeleteFromQueue known peer height
func (n *Node) deleteFromQueue(key string) {
	n.queue.sync.Lock()
	defer n.queue.sync.Unlock()

	delete(n.queue.heights, key)
}

// CheckQueue known peer height
func (n *Node) checkQueue(key string) bool {
	n.queue.sync.Lock()
	defer n.queue.sync.Unlock()

	_, ok := n.queue.heights[key]

	return ok
}

// Parent Queue

// addParentToQueue add parent event hash to queue
func (n *Node) addParentToQueue(key hash.Event) {
	n.parentQueue.sync.Lock()
	defer n.parentQueue.sync.Unlock()

	n.parentQueue.hashes[key] = true
}

// deleteParentFromQueue delete parent event hash from queue
func (n *Node) deleteParentFromQueue(key hash.Event) {
	n.parentQueue.sync.Lock()
	defer n.parentQueue.sync.Unlock()

	delete(n.parentQueue.hashes, key)
}

// checkParentQueue check parent event hash in queue
func (n *Node) checkParentQueue(key hash.Event) bool {
	n.parentQueue.sync.Lock()
	defer n.parentQueue.sync.Unlock()

	_, ok := n.parentQueue.hashes[key]

	return ok
}
