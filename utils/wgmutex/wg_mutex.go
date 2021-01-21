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
	m.RWMutex.Lock()
	m.wg.Wait()
}

func (m *WgMutex) RLock() {
	m.RWMutex.RLock()
	m.wg.Wait()
}
