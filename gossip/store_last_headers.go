package gossip

import (
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// DelLastHeader deletes record about last header from a validator
func (s *Store) DelLastHeader(epoch idx.Epoch, creator idx.StakerID) {
	key := append(epoch.Bytes(), creator.Bytes()...)

	err := s.table.LastEpochHeaders.Delete(key)
	if err != nil {
		s.Log.Crit("Failed to erase LastHeader", "err", err)
	}
}

// DelLastHeaders deletes all the records about last headers
func (s *Store) DelLastHeaders(epoch idx.Epoch) {
	it := s.table.LastEpochHeaders.NewIteratorWithPrefix(epoch.Bytes())
	defer it.Release()
	s.dropTable(it, s.table.LastEpochHeaders)
}

// AddLastHeader adds/updates a records about last header from a validator
func (s *Store) AddLastHeader(epoch idx.Epoch, header *inter.EventHeaderData) {
	key := append(epoch.Bytes(), header.Creator.Bytes()...)

	s.set(s.table.LastEpochHeaders, key, header)
}

// GetLastHeaders retrieves all the records about last headers from validators
func (s *Store) GetLastHeaders(epoch idx.Epoch) inter.HeadersByCreator {
	hh := make(inter.HeadersByCreator)

	it := s.table.LastEpochHeaders.NewIteratorWithPrefix(epoch.Bytes())
	defer it.Release()
	for it.Next() {
		creator := it.Key()[4:]
		header := &inter.EventHeaderData{}
		err := rlp.DecodeBytes(it.Value(), header)
		if err != nil {
			s.Log.Crit("Failed to decode rlp", "err", err)
		}
		hh[idx.BytesToStakerID(creator)] = header
	}

	return hh
}
