package utils

import (
	"sync"
)

type NumQueue struct {
	mu       sync.Mutex
	lastDone uint64
	waiters  []chan struct{}
}

func NewNumQueue(init uint64) *NumQueue {
	return &NumQueue{
		lastDone: init,
	}
}

func (q *NumQueue) Done(n uint64) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if n <= q.lastDone {
		panic("Already done!")
	}

	pos := int(n - q.lastDone - 1)
	for i := 0; i < len(q.waiters) && i <= pos; i++ {
		close(q.waiters[i])
	}
	if pos < len(q.waiters) {
		q.waiters = q.waiters[pos+1:]
	} else {
		q.waiters = make([]chan struct{}, 0, 1000)
	}

	q.lastDone = n
}

func (q *NumQueue) WaitFor(n uint64) {
	q.mu.Lock()

	if n <= q.lastDone {
		q.mu.Unlock()
		return
	}

	count := int(n - q.lastDone)
	for i := len(q.waiters); i < count; i++ {
		q.waiters = append(q.waiters, make(chan struct{}))
	}
	ch := q.waiters[count-1]
	q.mu.Unlock()
	<-ch
}
