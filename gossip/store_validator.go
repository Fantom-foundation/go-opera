package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
)

// AddBlocksMissed add count of missed blocks for validator
func (s *Store) IncBlocksMissed(v common.Address) {
	s.table.incMutex.Lock()
	defer s.table.incMutex.Unlock()

	missed := s.GetBlocksMissed(v)
	missed++
	err := s.table.BlockParticipation.Put(v.Bytes(), bigendian.Int32ToBytes(missed))
	if err != nil {
		s.Log.Crit("Failed to set key-value", "err", err)
	}
}

// ResetBlocksMissed set to 0 missed blocks for validator
func (s *Store) ResetBlocksMissed(v common.Address) {
	s.table.incMutex.Lock()
	defer s.table.incMutex.Unlock()

	err := s.table.BlockParticipation.Put(v.Bytes(), bigendian.Int32ToBytes(0))
	if err != nil {
		s.Log.Crit("Failed to set key-value", "err", err)
	}
}

// GetBlocksMissed return blocks missed num for validator
func (s *Store) GetBlocksMissed(v common.Address) uint32 {
	missedBytes, err := s.table.BlockParticipation.Get(v.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if missedBytes == nil {
		return 0
	}
	return bigendian.BytesToInt32(missedBytes)
}

// AddActiveValidatorsScore add gas value for active validation score
func (s *Store) AddActiveValidatorsScore(v common.Address, gas uint64) {
	s.addValidatorScore(s.table.ActiveValidatorScores, v, gas)
}

// GetActiveValidatorsScore return gas value for active validator score
func (s *Store) GetActiveValidatorsScore(v common.Address) uint64 {
	return s.getValidatorScore(s.table.ActiveValidatorScores, v)
}

// AddDirtyValidatorsScore add gas value for active validation score
func (s *Store) AddDirtyValidatorsScore(v common.Address, gas uint64) {
	s.addValidatorScore(s.table.DirtyValidatorScores, v, gas)
}

// GetDirtyValidatorsScore return gas value for active validator score
func (s *Store) GetDirtyValidatorsScore(v common.Address) uint64 {
	return s.getValidatorScore(s.table.DirtyValidatorScores, v)
}

func (s *Store) addValidatorScore(t kvdb.KeyValueStore, v common.Address, val uint64) {
	s.table.incMutex.Lock()
	defer s.table.incMutex.Unlock()

	score := s.getValidatorScore(t, v)
	score += val
	err := t.Put(v.Bytes(), bigendian.Int64ToBytes(score))
	if err != nil {
		s.Log.Crit("Failed to set key-value", "err", err)
	}
}

func (s *Store) getValidatorScore(t kvdb.KeyValueStore, v common.Address) uint64 {
	gasBytes, err := t.Get(v.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if gasBytes == nil {
		return 0
	}
	return bigendian.BytesToInt64(gasBytes)
}
