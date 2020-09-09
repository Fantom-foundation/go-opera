package gossip

import (
	"fmt"
	"math/big"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/sfctype"
	"github.com/Fantom-foundation/go-opera/opera"
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

// ApplyGenesis writes initial state.
func (s *Store) ApplyGenesis(net *opera.Config) (genesisAtropos hash.Event, genesisState common.Hash, new bool, err error) {
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

// calcGenesisHash calcs hash of genesis state.
func calcGenesisHash(net *opera.Config) hash.Event {
	s := NewMemStore()
	defer s.Close()

	h, _, _ := s.applyGenesis(net)

	return h
}

func (s *Store) applyGenesis(net *opera.Config) (genesisAtropos hash.Event, genesisState common.Hash, err error) {
	// apply app genesis
	state, err := s.app.ApplyGenesis(net)
	if err != nil {
		return genesisAtropos, genesisState, err
	}

	prettyHash := func(net *opera.Config) hash.Event {
		e := inter.MutableEventPayload{}
		// for nice-looking ID
		e.SetEpoch(0)
		e.SetLamport(idx.Lamport(net.Dag.MaxEpochBlocks))
		// actual data hashed
		h := net.Genesis.Alloc.Accounts.Hash()
		e.SetExtra(append(net.Genesis.ExtraData, h[:]...))
		e.SetCreationTime(net.Genesis.Time)

		return e.Build().ID()
	}
	genesisAtropos = prettyHash(net)
	genesisState = common.Hash(genesisAtropos)

	block := &inter.Block{
		Time:       net.Genesis.Time,
		Atropos:    genesisAtropos,
		Events:     hash.Events{genesisAtropos},
		SkippedTxs: []uint32{},
		GasUsed:    0,
		Root:       hash.Hash(state.Root),
	}

	s.SetBlock(0, block)
	s.SetBlockIndex(genesisAtropos, 0)

	validators := net.Genesis.Alloc.Validators.Build()
	valBlockStates := make([]ValidatorBlockState, validators.Len())
	for i := range valBlockStates {
		valBlockStates[i] = ValidatorBlockState{
			Originated: new(big.Int),
		}
	}
	valEpochStates := make([]ValidatorEpochState, validators.Len())

	valProfiles := make([]sfctype.SfcStakerAndID, len(net.Genesis.Alloc.Validators))
	for i, validator := range net.Genesis.Alloc.Validators {
		staker := &sfctype.SfcStaker{
			Address:      validator.Address,
			CreatedEpoch: 0,
			CreationTime: net.Genesis.Time,
			StakeAmount:  validator.Stake,
			DelegatedMe:  big.NewInt(0),
		}
		valProfiles[i] = sfctype.SfcStakerAndID{
			ValidatorID: validator.ID,
			Staker:      staker,
		}
		s.app.SetSfcStaker(validator.ID, staker)
	}

	s.SetBlockState(BlockState{
		Block:           0,
		EpochBlocks:     0,
		EpochFee:        new(big.Int),
		ValidatorStates: valBlockStates,
	})
	s.SetEpochState(EpochState{
		Epoch:             1,
		Validators:        validators,
		EpochStart:        net.Genesis.Time,
		PrevEpochStart:    net.Genesis.Time,
		ValidatorStates:   valEpochStates,
		ValidatorProfiles: valProfiles,
	})

	return genesisAtropos, genesisState, nil
}
