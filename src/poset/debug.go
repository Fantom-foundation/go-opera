// +build debug

// Package poset These functions are used only in debugging
package poset

import (
	"github.com/sirupsen/logrus"
)

// PrintStat prints the stats for logger
func (p *Poset) PrintStat(logger *logrus.Entry) {
	logger.Warn("****Known events:");
	for pidID, index := range p.Store.KnownEvents() {
		logger.Warn("    index=", index, " peer=", p.Participants.ByID[int64(pidID)].NetAddr,
			" pubKeyHex=", p.Participants.ByID[int64(pidID)].PubKeyHex)
	}
}

// TopologicalEvents returns all of badgers topological events (lamport)
func (s *BadgerStore) TopologicalEvents() ([]Event, error) {
	return s.dbTopologicalEvents()
}

// TopologicalEvents This is just a stub
func (s *InmemStore) TopologicalEvents() ([]Event, error) {
	return []Event{}, nil
}
