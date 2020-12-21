package wgmutex

import "sync"

type WgMutex struct {
	*sync.RWMutex
	wg *sync.WaitGroup
}

func New(m *sync.RWMutex, wg *sync.WaitGroup) *WgMutex {
	return &WgMutex{
		RWMutex: m,
		wg:      wg,
	}
}

func (m *WgMutex) Lock() {
	m.wg.Wait() // wait once before locking for better concurrency
	m.RWMutex.Lock()
	m.wg.Wait()
}

func (m *WgMutex) RLock() {
	m.wg.Wait() // wait once before locking for better concurrency
	m.RWMutex.RLock()
	m.wg.Wait()
}
