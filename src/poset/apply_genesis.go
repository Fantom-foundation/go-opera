package poset

import (
	"fmt"
	"github.com/Fantom-foundation/go-lachesis/src/common/bigendian"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis/genesis"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
)

// calcFirstGenesisHash calcs hash of genesis balances.
func calcFirstGenesisHash(config *genesis.Config) hash.Hash {
	s := NewMemStore()
	defer s.Close()

	_ = s.ApplyGenesis(config)

	return s.GetGenesis().PrevEpoch.Hash()
}

// ApplyGenesis stores initial state.
func (s *Store) ApplyGenesis(config *genesis.Config) error {
	if config == nil {
		return fmt.Errorf("config shouldn't be nil")
	}
	if config.Balances == nil {
		return fmt.Errorf("balances shouldn't be nil")
	}

	sf1 := s.GetGenesis()
	if sf1 != nil {
		if sf1.PrevEpoch.Hash() == calcFirstGenesisHash(config) {
			return nil
		}
		return fmt.Errorf("other genesis has applied already")
	}

	sf := &superFrame{}
	cp := &checkpoint{}

	sf.Members = make(pos.Members, len(config.Balances))
	for addr, balance := range config.Balances {
		if balance == 0 {
			return fmt.Errorf("balance shouldn't be zero")
		}

		sf.Members.Set(addr, balance)
	}
	sf.Members = sf.Members.Top()
	cp.NextMembers = sf.Members.Copy()

	// hash of NetworkId
	dummyFiWitness := inter.NewEvent()
	dummyFiWitness.Epoch = 0
	dummyFiWitness.Lamport = 1
	dummyFiWitness.Extra = bigendian.Int64ToBytes(config.NetworkId)

	// genesis object
	sf.SuperFrameN = firstEpoch
	sf.PrevEpoch.Epoch = sf.SuperFrameN - 1
	sf.PrevEpoch.StateHash = cp.StateHash
	sf.PrevEpoch.LastFiWitness = dummyFiWitness.Hash()
	sf.PrevEpoch.Time = config.Time
	cp.LastConsensusTime = sf.PrevEpoch.Time

	s.SetGenesis(sf)
	s.SetSuperFrame(sf)
	s.SetCheckpoint(cp)

	return nil
}
