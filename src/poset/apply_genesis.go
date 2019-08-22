package poset

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis/genesis"
)

// calcFirstGenesisHash calcs hash of genesis balances.
func calcFirstGenesisHash(g *genesis.Genesis, genesisFiWitness hash.Event, stateHash hash.Hash) hash.Hash {
	s := NewMemStore()
	defer s.Close()

	_ = s.ApplyGenesis(g, genesisFiWitness, stateHash)

	return s.GetGenesis().PrevEpoch.Hash()
}

// ApplyGenesis stores initial state.
func (s *Store) ApplyGenesis(g *genesis.Genesis, genesisFiWitness hash.Event, stateHash hash.Hash) error {
	if g == nil {
		return fmt.Errorf("config shouldn't be nil")
	}
	if g.Alloc == nil {
		return fmt.Errorf("balances shouldn't be nil")
	}

	sf1 := s.GetGenesis()
	if sf1 != nil {
		if sf1.PrevEpoch.Hash() == calcFirstGenesisHash(g, genesisFiWitness, stateHash) {
			return nil
		}
		return fmt.Errorf("other genesis has applied already")
	}

	sf := &superFrame{}
	cp := &checkpoint{
		StateHash: stateHash,
	}

	sf.Members = make(pos.Members, len(g.Alloc))
	/*for addr, account := range g.Alloc {
		if account == 0 {
			return fmt.Errorf("balance shouldn't be zero")
		}

		sf.Members.Set(addr, account.Balance) TODO
	}*/
	sf.Members = sf.Members.Top()
	cp.NextMembers = sf.Members.Copy()

	// genesis object
	sf.SuperFrameN = firstEpoch
	sf.PrevEpoch.Epoch = sf.SuperFrameN - 1
	sf.PrevEpoch.StateHash = cp.StateHash
	sf.PrevEpoch.LastFiWitness = genesisFiWitness
	sf.PrevEpoch.Time = g.Time
	cp.LastConsensusTime = sf.PrevEpoch.Time
	cp.LastFiWitness = genesisFiWitness

	s.SetGenesis(sf)
	s.SetSuperFrame(sf)
	s.SetCheckpoint(cp)

	return nil
}
