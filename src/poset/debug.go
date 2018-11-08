// +build debug

// These functions are used only in debugging
package poset

import (
	"fmt"
)

func (p *Poset) PrintStat() {
	fmt.Println("****Known events:");
	for pid_id, index := range p.Store.KnownEvents() {
		fmt.Println("    index=", index, " peer=", p.Participants.ById[int64(pid_id)].NetAddr,
			" pubKeyHex=", p.Participants.ById[int64(pid_id)].PubKeyHex)
	}
}

func (s *BadgerStore) TopologicalEvents() ([]Event, error) {
	return s.dbTopologicalEvents()
}

// This is just a stub
func (s *InmemStore) TopologicalEvents() ([]Event, error) {
	return []Event{}, nil
}

