package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// SetBlockDecidedBy stores event which decided block (off-chain data)
func (s *Store) SetBlockDecidedBy(block idx.Block, event hash.Event) {
	err := s.table.DecisiveEvents.Put(block.Bytes(), event.Bytes())
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
}

// GetBlockDecidedBy get event which decided block (off-chain data)
func (s *Store) GetBlockDecidedBy(block idx.Block) hash.Event {
	idBytes, err := s.table.DecisiveEvents.Get(block.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}
	if idBytes == nil {
		return hash.Event{}
	}

	return hash.BytesToEvent(idBytes)
}

// SetEventReceivingTime stores local event time (off-chain data)
func (s *Store) SetEventReceivingTime(e hash.Event, time inter.Timestamp) {
	err := s.table.EventLocalTimes.Put(e.Bytes(), time.Bytes())
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
}

// GetEventReceivingTime get local event time (off-chain data)
func (s *Store) GetEventReceivingTime(e hash.Event) inter.Timestamp {
	timeBytes, err := s.table.EventLocalTimes.Get(e.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}
	if timeBytes == nil {
		return 0
	}

	return inter.BytesToTimestamp(timeBytes)
}
