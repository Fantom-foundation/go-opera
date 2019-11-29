package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
)

const (
	ValidatorsScoreCheckpointKey = "LastScoreCheckpoint"
)

// IncBlocksMissed add count of missed blocks for validator
func (s *Store) IncBlocksMissed(stakerID idx.StakerID) {
	s.mutexes.IncMutex.Lock()
	defer s.mutexes.IncMutex.Unlock()

	missed := s.GetBlocksMissed(stakerID)
	missed++
	err := s.table.BlockParticipation.Put(stakerID.Bytes(), bigendian.Int32ToBytes(missed))
	if err != nil {
		s.Log.Crit("Failed to set key-value", "err", err)
	}

	s.cache.BlockParticipation.Add(stakerID, missed)
}

// ResetBlocksMissed set to 0 missed blocks for validator
func (s *Store) ResetBlocksMissed(stakerID idx.StakerID) {
	s.mutexes.IncMutex.Lock()
	defer s.mutexes.IncMutex.Unlock()

	err := s.table.BlockParticipation.Delete(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to set key-value", "err", err)
	}

	s.cache.BlockParticipation.Add(stakerID, uint32(0))
}

// GetBlocksMissed return blocks missed num for validator
func (s *Store) GetBlocksMissed(stakerID idx.StakerID) uint32 {
	missedVal, ok := s.cache.BlockParticipation.Get(stakerID)
	if ok {
		missed, ok := missedVal.(uint32)
		if ok {
			return missed
		}
	}

	missedBytes, err := s.table.BlockParticipation.Get(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if missedBytes == nil {
		return 0
	}

	missed := bigendian.BytesToInt32(missedBytes)
	s.cache.BlockParticipation.Add(stakerID, missed)

	return missed
}

// AddActiveValidatorsScore add gas value for active validation score
func (s *Store) AddActiveValidatorsScore(stakerID idx.StakerID, gas uint64) {
	s.addValidatorScore(s.table.ActiveValidatorScores, stakerID, gas)
}

// GetActiveValidatorsScore return gas value for active validator score
func (s *Store) GetActiveValidatorsScore(stakerID idx.StakerID) uint64 {
	return s.getValidatorScore(s.table.ActiveValidatorScores, stakerID)
}

// AddDirtyValidatorsScore add gas value for active validation score
func (s *Store) AddDirtyValidatorsScore(stakerID idx.StakerID, gas uint64) {
	s.addValidatorScore(s.table.DirtyValidatorScores, stakerID, gas)
}

func (s *Store) DelActiveValidatorsScore(stakerID idx.StakerID) {
	err := s.table.ActiveValidatorScores.Delete(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase key-value", "err", err)
	}
}

func (s *Store) DelDirtyValidatorsScore(stakerID idx.StakerID) {
	err := s.table.DirtyValidatorScores.Delete(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase key-value", "err", err)
	}
}

// GetDirtyValidatorsScore return gas value for active validator score
func (s *Store) GetDirtyValidatorsScore(stakerID idx.StakerID) uint64 {
	return s.getValidatorScore(s.table.DirtyValidatorScores, stakerID)
}

func (s *Store) addValidatorScore(t kvdb.KeyValueStore, stakerID idx.StakerID, val uint64) {
	s.mutexes.IncMutex.Lock()
	defer s.mutexes.IncMutex.Unlock()

	score := s.getValidatorScore(t, stakerID)
	score += val
	err := t.Put(stakerID.Bytes(), bigendian.Int64ToBytes(score))
	if err != nil {
		s.Log.Crit("Failed to set key-value", "err", err)
	}
}

func (s *Store) getValidatorScore(t kvdb.KeyValueStore, stakerID idx.StakerID) uint64 {
	gasBytes, err := t.Get(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if gasBytes == nil {
		return 0
	}
	return bigendian.BytesToInt64(gasBytes)
}

// SetScoreCheckpoint set score checkpoint time
func (s *Store) SetScoreCheckpoint(cp inter.Timestamp) {
	cpBytes := bigendian.Int64ToBytes(uint64(cp))
	err := s.table.ScoreCheckpoint.Put([]byte(ValidatorsScoreCheckpointKey), cpBytes)
	if err != nil {
		s.Log.Crit("Failed to set key-value", "err", err)
	}

	s.cache.BlockParticipation.Add(ValidatorsScoreCheckpointKey, cp)
}

// GetScoreCheckpoint return last score checkpoint time
func (s *Store) GetScoreCheckpoint() inter.Timestamp {
	cpVal, ok := s.cache.ScoreCheckpoint.Get(ValidatorsScoreCheckpointKey)
	if ok {
		cp, ok := cpVal.(inter.Timestamp)
		if ok {
			return cp
		}
	}

	cpBytes, err := s.table.ScoreCheckpoint.Get([]byte(ValidatorsScoreCheckpointKey))
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if cpBytes == nil {
		return 0
	}

	cp := inter.Timestamp(bigendian.BytesToInt64(cpBytes))
	s.cache.BlockParticipation.Add(ValidatorsScoreCheckpointKey, cp)

	return cp
}

func (s *Store) MoveDirtyValidatorsToActive() {
	it := s.table.DirtyValidatorScores.NewIterator()
	defer it.Release()

	keys := make([][]byte, 0, 500) // don't write during iteration
	vals := make([][]byte, 0, 500)

	for it.Next() {
		keys = append(keys, it.Key())
		vals = append(keys, it.Value())
	}

	for i := range keys {
		err := s.table.ActiveValidatorScores.Put(keys[i], vals[i])
		if err != nil {
			s.Log.Crit("Failed to set key-value", "err", err)
		}
		err = s.table.DirtyValidatorScores.Delete(keys[i])
		if err != nil {
			s.Log.Crit("Failed to erase key-value", "err", err)
		}
	}
}
