package reporter

import (
	"sync"
)

type SingleStack struct {
	lock sync.Mutex
	job  *WebSocketJob
}

func NewSingleStack() *SingleStack {
	return &SingleStack{sync.Mutex{}, nil}
}

func (s *SingleStack) Push(v *WebSocketJob) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.job = v
}

func (s *SingleStack) Pop() *WebSocketJob {
	s.lock.Lock()
	defer s.lock.Unlock()

	v := s.job
	s.job = nil
	return v
}
