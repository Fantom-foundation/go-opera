package gossip

import (
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/inter/pos"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/inter/sfctype"
)

const lbKey = "d"

const leKey = "e"

// SetBlockState stores the latest block state
func (s *Store) SetBlockState(v BlockState) {
	s.set(s.table.BlockState, []byte(lbKey), v)
	s.cache.BlockState.Store(&v)
}

// GetBlockState retrieves the latest block state
func (s *Store) GetBlockState() BlockState {
	if v := s.cache.BlockState.Load(); v != nil {
		return *v.(*BlockState)
	}
	v, ok := s.get(s.table.BlockState, []byte(lbKey), &BlockState{}).(*BlockState)
	if !ok {
		log.Crit("Genesis not applied")
	}
	s.cache.BlockState.Store(v)
	return *v
}

// SetEpochState stores the latest block state
func (s *Store) SetEpochState(v EpochState) {
	s.set(s.table.EpochState, []byte(leKey), v)
	s.cache.EpochState.Store(&v)
}

// GetEpochState retrieves the latest block state
func (s *Store) GetEpochState() EpochState {
	if v := s.cache.EpochState.Load(); v != nil {
		return *v.(*EpochState)
	}
	v, ok := s.get(s.table.EpochState, []byte(leKey), &EpochState{}).(*EpochState)
	if !ok {
		log.Crit("Genesis not applied")
	}
	s.cache.EpochState.Store(v)
	return *v
}

// GetEpoch retrieves the current epoch
func (s *Store) GetEpoch() idx.Epoch {
	return s.GetEpochState().Epoch
}

// GetValidators retrieves alidators atomically
func (s *Store) GetValidators() *pos.Validators {
	return s.GetEpochState().Validators
}

// GetEpoch retrieves the current epoch and validators atomically
func (s *Store) GetEpochValidators() (*pos.Validators, idx.Epoch) {
	state := s.GetEpochState()
	return state.Validators, state.Epoch
}

// GetEpoch retrieves the current block number
func (s *Store) GetLatestBlockIndex() idx.Block {
	return s.GetBlockState().Block
}

// GetEpochBlockIndex retrieves the number of blocks in current epoch
func (s *Store) GetEpochBlocks() idx.Block {
	return s.GetBlockState().EpochBlocks
}

func (s *Store) GetValidatorProfiles() []sfctype.SfcStakerAndID {
	return s.GetEpochState().ValidatorProfiles
}
