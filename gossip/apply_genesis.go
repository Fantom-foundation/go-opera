package gossip

import (
	"fmt"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/ethereum/go-ethereum/common"
	"math/big"

	"github.com/Fantom-foundation/go-lachesis/evmcore"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/sfctype"
	"github.com/Fantom-foundation/go-lachesis/lachesis"
)

// GenesisMismatchError is raised when trying to overwrite an existing
// genesis block with an incompatible one.
type GenesisMismatchError struct {
	Stored, New hash.Event
}

// Error implements error interface.
func (e *GenesisMismatchError) Error() string {
	return fmt.Sprintf("database contains incompatible gossip genesis (have %s, new %s)", e.Stored.FullID(), e.New.FullID())
}

// calcGenesisHash calcs hash of genesis state.
func calcGenesisHash(net *lachesis.Config) hash.Event {
	s := NewMemStore()
	defer s.Close()

	h, _, _ := s.applyGenesis(net)

	return h
}

// ApplyGenesis writes initial state.
func (s *Store) ApplyGenesis(net *lachesis.Config) (genesisAtropos hash.Event, genesisState common.Hash, new bool, err error) {
	storedGenesis := s.GetBlock(0)
	if storedGenesis != nil {
		newHash := calcGenesisHash(net)
		if storedGenesis.Atropos != newHash {
			return genesisAtropos, genesisState, true, &GenesisMismatchError{storedGenesis.Atropos, newHash}
		}

		genesisAtropos = storedGenesis.Atropos
		genesisState = common.Hash(genesisAtropos)
		return genesisAtropos, genesisState, false, nil
	}
	// if we'here, then it's first time genesis is applied
	genesisAtropos, genesisState, err = s.applyGenesis(net)
	if err != nil {
		return genesisAtropos, genesisState, true, err
	}

	return genesisAtropos, genesisState, true, err
}

func (s *Store) applyGenesis(net *lachesis.Config) (genesisAtropos hash.Event, genesisState common.Hash, err error) {
	evmBlock, err := evmcore.ApplyGenesis(s.table.Evm, net)
	if err != nil {
		return genesisAtropos, genesisState, err
	}

	prettyHash := func(net *lachesis.Config) hash.Event {
		e := inter.NewEvent()
		// for nice-looking ID
		e.Epoch = 0
		e.Lamport = idx.Lamport(net.Dag.MaxEpochBlocks)
		// actual data hashed
		e.Extra = net.Genesis.ExtraData
		e.ClaimedTime = net.Genesis.Time
		e.TxHash = net.Genesis.Alloc.Accounts.Hash()

		return e.CalcHash()
	}
	genesisAtropos = prettyHash(net)
	genesisState = common.Hash(genesisAtropos)

	block := inter.NewBlock(0,
		net.Genesis.Time,
		genesisAtropos,
		hash.Event{},
		hash.Events{genesisAtropos},
	)

	block.Root = evmBlock.Root
	s.SetBlock(block)
	s.SetBlockIndex(genesisAtropos, block.Index)
	s.SetEpochStats(0, &sfctype.EpochStats{
		Start:    net.Genesis.Time,
		End:      net.Genesis.Time,
		TotalFee: new(big.Int),
	})
	s.SetDirtyEpochStats(&sfctype.EpochStats{
		Start:    net.Genesis.Time,
		TotalFee: new(big.Int),
	})

	for _, validator := range net.Genesis.Alloc.Validators {
		staker := &sfctype.SfcStaker{
			Address:      validator.Address,
			CreatedEpoch: 0,
			CreatedTime:  net.Genesis.Time,
			StakeAmount:  validator.Stake,
			DelegatedMe:  big.NewInt(0),
		}
		s.SetSfcStaker(validator.ID, staker)
		s.SetEpochValidator(1, validator.ID, staker)
	}

	return genesisAtropos, genesisState, nil
}
