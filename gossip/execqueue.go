package gossip

import "sync"

// execQueue implements a queue that executes function calls in a single thread,
// in the same order as they have been queued.
type execQueue struct {
	mu        sync.Mutex
	cond      *sync.Cond
	funcs     []func()
	closeWait chan struct{}
}

// newExecQueue creates a new execution queue.
func newExecQueue(capacity int) *execQueue {
	q := &execQueue{funcs: make([]func(), 0, capacity)}
	q.cond = sync.NewCond(&q.mu)
	go q.loop()
	return q
}

func (q *execQueue) loop() {
	for f := q.waitNext(false); f != nil; f = q.waitNext(true) {
		f()
	}
	close(q.closeWait)
}

func (q *execQueue) waitNext(drop bool) (f func()) {
	q.mu.Lock()
	if drop && len(q.funcs) > 0 {
		// Remove the function that just executed. We do this here instead of when
		// dequeuing so len(q.funcs) includes the function that is running.
		q.funcs = append(q.funcs[:0], q.funcs[1:]...)
	}
	for !q.isClosed() {
		if len(q.funcs) > 0 {
			f = q.funcs[0]
			break
		}
		q.cond.Wait()
	}
	q.mu.Unlock()
	return f
}

func (q *execQueue) isClosed() bool {
	return q.closeWait != nil
}

// canQueue returns true if more function calls can be added to the execution queue.
func (q *execQueue) canQueue() bool {
	q.mu.Lock()
	ok := !q.isClosed() && len(q.funcs) < cap(q.funcs)
	q.mu.Unlock()
	return ok
}

// queue adds a function call to the execution queue. Returns true if successful.
func (q *execQueue) queue(f func()) bool {
	q.mu.Lock()
	ok := !q.isClosed() && len(q.funcs) < cap(q.funcs)
	if ok {
		q.funcs = append(q.funcs, f)
		q.cond.Signal()
	}
	q.mu.Unlock()
	return ok
}

// clear drops all queued functions
func (q *execQueue) clear() {
	q.mu.Lock()
	q.funcs = q.funcs[:0]
	q.mu.Unlock()
}

// quit stops the exec queue.
// quit waits for the current execution to finish before returning.
func (q *execQueue) quit() {
	q.mu.Lock()
	if !q.isClosed() {
		q.closeWait = make(chan struct{})
		q.cond.Signal()
	}
	q.mu.Unlock()
	<-q.closeWait
}
