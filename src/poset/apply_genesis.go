package poset

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis/genesis"
)

// calcFirstGenesisHash calcs hash of genesis balances.
func calcFirstGenesisHash(g *genesis.Genesis, genesisAtropos hash.Event, stateHash common.Hash) common.Hash {
	s := NewMemStore()
	defer s.Close()

	_ = s.ApplyGenesis(g, genesisAtropos, stateHash)

	return s.GetGenesis().PrevEpoch.Hash()
}

// ApplyGenesis stores initial state.
func (s *Store) ApplyGenesis(g *genesis.Genesis, genesisAtropos hash.Event, stateHash common.Hash) error {
	if g == nil {
		return fmt.Errorf("config shouldn't be nil")
	}
	if g.Alloc == nil {
		return fmt.Errorf("balances shouldn't be nil")
	}

	if exist := s.GetGenesis(); exist != nil {
		if exist.PrevEpoch.Hash() == calcFirstGenesisHash(g, genesisAtropos, stateHash) {
			return nil
		}
		return fmt.Errorf("other genesis has applied already")
	}

	e := &epoch{}
	cp := &checkpoint{
		StateHash: stateHash,
	}

	e.Members = make(pos.Members, len(g.Alloc))
	for addr, account := range g.Alloc {
		e.Members.Set(addr, pos.BalanceToStake(account.Balance))
	}
	e.Members = e.Members.Top()
	cp.NextMembers = e.Members.Copy()

	// genesis object
	e.EpochN = firstEpoch
	e.PrevEpoch.Epoch = e.EpochN - 1
	e.PrevEpoch.StateHash = cp.StateHash
	e.PrevEpoch.LastAtropos = genesisAtropos
	e.PrevEpoch.Time = g.Time
	cp.LastConsensusTime = e.PrevEpoch.Time
	cp.LastAtropos = genesisAtropos

	s.SetGenesis(e)
	s.SetEpoch(e)
	s.SetCheckpoint(cp)

	return nil
}
