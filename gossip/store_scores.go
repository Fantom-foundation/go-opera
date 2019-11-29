package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
)

const (
	ValidationScoreCheckpointKey = "LastScoreCheckpoint"
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

// AddActiveValidationScore add gas value for active validation score
func (s *Store) AddActiveValidationScore(stakerID idx.StakerID, gas uint64) {
	s.addValidationScore(s.table.ActiveValidationScore, stakerID, gas)
}

// GetActiveValidationScore return gas value for active validator score
func (s *Store) GetActiveValidationScore(stakerID idx.StakerID) uint64 {
	return s.getValidationScore(s.table.ActiveValidationScore, stakerID)
}

// AddDirtyValidationScore add gas value for active validation score
func (s *Store) AddDirtyValidationScore(stakerID idx.StakerID, gas uint64) {
	s.addValidationScore(s.table.DirtyValidationScore, stakerID, gas)
}

func (s *Store) DelActiveValidationScore(stakerID idx.StakerID) {
	err := s.table.ActiveValidationScore.Delete(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase key-value", "err", err)
	}
}

func (s *Store) DelDirtyValidationScore(stakerID idx.StakerID) {
	err := s.table.DirtyValidationScore.Delete(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase key-value", "err", err)
	}
}

// GetDirtyValidationScore return gas value for active validator score
func (s *Store) GetDirtyValidationScore(stakerID idx.StakerID) uint64 {
	return s.getValidationScore(s.table.DirtyValidationScore, stakerID)
}

func (s *Store) addValidationScore(t kvdb.KeyValueStore, stakerID idx.StakerID, val uint64) {
	s.mutexes.IncMutex.Lock()
	defer s.mutexes.IncMutex.Unlock()

	score := s.getValidationScore(t, stakerID)
	score += val
	err := t.Put(stakerID.Bytes(), bigendian.Int64ToBytes(score))
	if err != nil {
		s.Log.Crit("Failed to set key-value", "err", err)
	}
}

func (s *Store) getValidationScore(t kvdb.KeyValueStore, stakerID idx.StakerID) uint64 {
	scoreBytes, err := t.Get(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if scoreBytes == nil {
		return 0
	}
	return bigendian.BytesToInt64(scoreBytes)
}

// SetScoreCheckpoint set score checkpoint time
func (s *Store) SetScoreCheckpoint(cp inter.Timestamp) {
	cpBytes := bigendian.Int64ToBytes(uint64(cp))
	err := s.table.ScoreCheckpoint.Put([]byte(ValidationScoreCheckpointKey), cpBytes)
	if err != nil {
		s.Log.Crit("Failed to set key-value", "err", err)
	}

	s.cache.ScoreCheckpoint.Add(ValidationScoreCheckpointKey, cp)
}

// GetScoreCheckpoint return last score checkpoint time
func (s *Store) GetScoreCheckpoint() inter.Timestamp {
	cpVal, ok := s.cache.ScoreCheckpoint.Get(ValidationScoreCheckpointKey)
	if ok {
		cp, ok := cpVal.(inter.Timestamp)
		if ok {
			return cp
		}
	}

	cpBytes, err := s.table.ScoreCheckpoint.Get([]byte(ValidationScoreCheckpointKey))
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if cpBytes == nil {
		return 0
	}

	cp := inter.Timestamp(bigendian.BytesToInt64(cpBytes))
	s.cache.ScoreCheckpoint.Add(ValidationScoreCheckpointKey, cp)

	return cp
}

func (s *Store) MoveDirtyValidatorsToActive() {
	it := s.table.DirtyValidationScore.NewIterator()
	defer it.Release()

	keys := make([][]byte, 0, 500) // don't write during iteration
	vals := make([][]byte, 0, 500)

	for it.Next() {
		keys = append(keys, it.Key())
		vals = append(keys, it.Value())
	}

	for i := range keys {
		err := s.table.ActiveValidationScore.Put(keys[i], vals[i])
		if err != nil {
			s.Log.Crit("Failed to set key-value", "err", err)
		}
		err = s.table.DirtyValidationScore.Delete(keys[i])
		if err != nil {
			s.Log.Crit("Failed to erase key-value", "err", err)
		}
	}
}
