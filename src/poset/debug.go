// +build debug

// These functions are used only in debugging
package poset

import (
	"fmt"
)

func (p *Poset) PrintStat() {
	fmt.Println("****Known events:");
	for pid_id, index := range p.Store.KnownEvents() {
		fmt.Println("    pid.ID=", pid_id, " index=", index)
	}
}

func (s *BadgerStore) TopologicalEvents() ([]Event, error) {
	return s.dbTopologicalEvents()
}

// This is just a stub
func (s *InmemStore) TopologicalEvents() ([]Event, error) {
	return []Event{}, nil
}

