package poset

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
)

// calcFirstGenesisHash calcs hash of genesis balances.
func calcFirstGenesisHash(g *lachesis.Genesis) hash.Hash {
	s := NewMemStore()
	defer s.Close()

	_ = s.ApplyGenesis(g)

	return s.GetGenesis().PrevEpoch.Hash()
}

// ApplyGenesis stores initial state.
func (s *Store) ApplyGenesis(g *lachesis.Genesis) error {
	if g == nil {
		return fmt.Errorf("config shouldn't be nil")
	}
	if g.Balances == nil {
		return fmt.Errorf("balances shouldn't be nil")
	}

	sf1 := s.GetGenesis()
	if sf1 != nil {
		if sf1.PrevEpoch.Hash() == calcFirstGenesisHash(g) {
			return nil
		}
		return fmt.Errorf("other genesis has applied already")
	}

	sf := &superFrame{}
	cp := &checkpoint{}

	sf.Members = make(pos.Members, len(g.Balances))
	for addr, balance := range g.Balances {
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

	// genesis object
	sf.SuperFrameN = firstEpoch
	sf.PrevEpoch.Epoch = sf.SuperFrameN - 1
	sf.PrevEpoch.StateHash = cp.StateHash
	sf.PrevEpoch.LastFiWitness = dummyFiWitness.Hash()
	sf.PrevEpoch.Time = g.Time
	cp.LastConsensusTime = sf.PrevEpoch.Time
	cp.LastFiWitness = dummyFiWitness.Hash()

	s.SetGenesis(sf)
	s.SetSuperFrame(sf)
	s.SetCheckpoint(cp)

	return nil
}
