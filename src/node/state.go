package node

import (
	"sync"
)

const (
	// Gossiping is the initial state of a Lachesis node.
	Gossiping state = iota
	// CatchingUp is the fast forward state
	CatchingUp
	// Shutdown is the shut down state
	Shutdown
	// Stop is the stop communicating state
	Stop
)

type state int

type nodeState2 struct {
	cond *sync.Cond
	lock sync.RWMutex
	wip  int

	state        state
	getStateChan chan state
	setStateChan chan state
}

func newNodeState2() *nodeState2 {
	ns := &nodeState2{
		cond:         sync.NewCond(&sync.Mutex{}),
		getStateChan: make(chan state),
		setStateChan: make(chan state),
	}
	go ns.mtx()
	return ns
}

func (s state) String() string {
	switch s {
	case Gossiping:
		return "Gossiping"
	case CatchingUp:
		return "CatchingUp"
	case Shutdown:
		return "Shutdown"
	case Stop:
		return "Stop"
	default:
		return "Unknown"
	}
}

func (s *nodeState2) mtx() {
	for {
		select {
		case s.state = <-s.setStateChan:
		case s.getStateChan <- s.state:
		}
	}
}

func (s *nodeState2) goFunc(fu func()) {
	go func() {
		s.lock.Lock()
		s.wip++
		s.lock.Unlock()

		fu()

		s.lock.Lock()
		s.wip--
		s.lock.Unlock()

		s.cond.L.Lock()
		defer s.cond.L.Unlock()
		s.cond.Broadcast()
	}()
}

func (s *nodeState2) waitRoutines() {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()

	for {
		s.lock.RLock()
		wip := s.wip
		s.lock.RUnlock()

		if wip != 0 {
			s.cond.Wait()
			continue
		}
		break
	}
}

func (s *nodeState2) getState() state {
	return <-s.getStateChan
}

func (s *nodeState2) setState(state state) {
	s.setStateChan <- state
}
