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

// GenesisMismatchError is raised when trying to overwrite an existing
// genesis block with an incompatible one.
type GenesisMismatchError struct {
	Stored, New common.Hash
}

// Error implements error interface.
func (e *GenesisMismatchError) Error() string {
	return fmt.Sprintf("database contains incompatible poset genesis (have %s, new %s)", e.Stored.String(), e.New.String())
}

// GenesisState stores state of previous Epoch
type GenesisState struct {
	Epoch       idx.Epoch
	Time        inter.Timestamp // consensus time of the last Atropos
	LastAtropos hash.Event
	AppHash     common.Hash
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
func calcGenesisHash(g *genesis.Genesis, genesisAtropos hash.Event, appHash common.Hash) common.Hash {
	s := NewMemStore()
	defer s.Close()

	_ = s.ApplyGenesis(g, genesisAtropos, appHash)

	return s.GetGenesis().PrevEpoch.Hash()
}

// ApplyGenesis writes initial state.
func (s *Store) ApplyGenesis(g *genesis.Genesis, genesisAtropos hash.Event, appHash common.Hash) error {
	if g == nil {
		return fmt.Errorf("genesis config shouldn't be nil")
	}
	if len(g.Alloc.Validators) == 0 {
		return fmt.Errorf("genesis validators shouldn't be empty")
	}

	if stored := s.GetGenesis(); stored != nil {
		storedHash := stored.PrevEpoch.Hash()
		newHash := calcGenesisHash(g, genesisAtropos, appHash)
		if storedHash != newHash {
			return &GenesisMismatchError{newHash, storedHash}
		}
		return nil
	}

	e := &EpochState{}
	cp := &Checkpoint{
		AppHash: appHash,
	}

	e.Validators = g.Alloc.Validators.Validators().Copy()

	// genesis object
	e.EpochN = firstEpoch
	e.PrevEpoch.Epoch = e.EpochN - 1
	e.PrevEpoch.AppHash = cp.AppHash
	e.PrevEpoch.LastAtropos = genesisAtropos
	e.PrevEpoch.Time = g.Time
	cp.LastAtropos = genesisAtropos

	s.SetGenesis(e)
	s.SetEpoch(e)
	s.SetCheckpoint(cp)

	return nil
}
