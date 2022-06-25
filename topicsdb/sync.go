package topicsdb

import (
	"sync"
)

type synchronizator struct {
	sync.Mutex
	threads   sync.WaitGroup
	positions []int
}

func newSynchronizator() *synchronizator {
	return &synchronizator{
		positions: make([]int, 0),
	}
}

func (s *synchronizator) StartThread(pos int, num int) {
	s.threads.Add(1)
	if len(s.positions) == 0 || s.positions[len(s.positions)-1] != pos {
		s.positions = append(s.positions, pos)
	}
}

func (s *synchronizator) FinishThread(pos int, num int) {
	s.threads.Done()
}

func (s *synchronizator) WaitForThreads() {
	s.threads.Wait()
}

func (s *synchronizator) Criteries() int {
	return len(s.positions)
}
