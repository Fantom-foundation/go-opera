package poset

import (
	"fmt"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis/genesis"
)

// state of previous Epoch
type GenesisState struct {
	Epoch       idx.Epoch
	Time        inter.Timestamp // consensus time of the last atropos
	LastAtropos hash.Event
	StateHash   common.Hash // hash of txs state
	LastHeaders headersByCreator
}

func (g *GenesisState) Hash() common.Hash {
	hasher := sha3.New256()
	if err := rlp.Encode(hasher, g); err != nil {
		panic(err)
	}
	return hash.FromBytes(hasher.Sum(nil))
}

func (g *GenesisState) EpochName() string {
	return fmt.Sprintf("epoch%d", g.Epoch)
}

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

	e := &epochState{}
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
	e.PrevEpoch.LastHeaders = headersByCreator{}
	cp.LastConsensusTime = e.PrevEpoch.Time
	cp.LastAtropos = genesisAtropos

	s.SetGenesis(e)
	s.SetEpoch(e)
	s.SetCheckpoint(cp)

	return nil
}
