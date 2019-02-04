package posposet

import (
	"fmt"
	"sync"
	"time"
)

// Poset is the main package struct.
type Poset struct {
	Store *Store

	Creator Node

	processing_wg sync.WaitGroup
	processing_ch chan struct{}
}

// New creates Poset instance.
func New(store *Store, key PrivateKey) *Poset {
	pk := key.PublicKey()
	return &Poset{
		Store: store,
		Creator: Node{
			ID:     AddressOf(pk),
			PubKey: pk,
			key:    &key,
		},
	}
}

// Start starts events processing. It is not thread-safe.
func (p *Poset) Start() {
	if p.processing_ch != nil {
		return
	}
	p.processing_ch = make(chan struct{})
	p.processing_wg.Add(1)
	go func() {
		defer p.processing_wg.Done()
		fmt.Println("start processing ...")
		for {
			select {
			case <-p.processing_ch:
				fmt.Println("stop processing ...")
				return
			case <-time.After(1 * time.Microsecond):
				fmt.Println("processing ...")
			}
		}
	}()
}

// Stop stops events processing. It is not thread-safe.
func (p *Poset) Stop() {
	if p.processing_ch == nil {
		return
	}
	close(p.processing_ch)
	p.processing_wg.Wait()
	p.processing_ch = nil
}

func (p *Poset) PushEvent(e Event) {

}
