package poset

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/lachesis/genesis"
)

// GenesisState stores state of previous Epoch
type GenesisState struct {
	Epoch       idx.Epoch
	Time        inter.Timestamp // consensus time of the last Atropos
	LastAtropos hash.Event
	StateHash   common.Hash // hash of txs state
	LastHeaders headersByCreator
}

func (g *GenesisState) Hash() common.Hash {
	hasher := sha3.NewLegacyKeccak256()
	if err := rlp.Encode(hasher, g); err != nil {
		panic(err)
	}
	return hash.FromBytes(hasher.Sum(nil))
}

func (g *GenesisState) EpochName() string {
	return fmt.Sprintf("epoch%d", g.Epoch)
}

// calcGenesisHash calcs hash of genesis state.
func calcGenesisHash(g *genesis.Genesis, genesisAtropos hash.Event, stateHash common.Hash) common.Hash {
	s := NewMemStore()
	defer s.Close()

	_ = s.ApplyGenesis(g, genesisAtropos, stateHash)

	return s.GetGenesis().PrevEpoch.Hash()
}

// ApplyGenesis stores initial state.
func (s *Store) ApplyGenesis(g *genesis.Genesis, genesisAtropos hash.Event, stateHash common.Hash) error {
	if g == nil {
		return fmt.Errorf("genesis config shouldn't be nil")
	}
	if g.Validators.Len() == 0 {
		return fmt.Errorf("genesis validators shouldn't be empty")
	}

	if exist := s.GetGenesis(); exist != nil {
		if exist.PrevEpoch.Hash() == calcGenesisHash(g, genesisAtropos, stateHash) {
			return nil
		}
		return fmt.Errorf("other genesis was applied already")
	}

	e := &EpochState{}
	cp := &Checkpoint{
		StateHash: stateHash,
	}

	e.Validators = g.Validators.Copy()

	// genesis object
	e.EpochN = firstEpoch
	e.PrevEpoch.Epoch = e.EpochN - 1
	e.PrevEpoch.StateHash = cp.StateHash
	e.PrevEpoch.LastAtropos = genesisAtropos
	e.PrevEpoch.Time = g.Time
	e.PrevEpoch.LastHeaders = headersByCreator{}
	cp.LastAtropos = genesisAtropos

	s.SetGenesis(e)
	s.SetEpoch(e)
	s.SetCheckpoint(cp)

	return nil
}
