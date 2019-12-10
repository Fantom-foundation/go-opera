package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
)

const (
	validationScoreCheckpointKey  = "current"
	originationScoreCheckpointKey = "current"
)

// IncBlocksMissed add count of missed blocks for validator
func (s *Store) IncBlocksMissed(stakerID idx.StakerID, periodDiff inter.Timestamp) {
	s.mutexes.IncMutex.Lock()
	defer s.mutexes.IncMutex.Unlock()

	missed := s.GetBlocksMissed(stakerID)
	missed.Num++
	missed.Period += periodDiff
	s.set(s.table.BlockParticipation, stakerID.Bytes(), &missed)

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

	s.cache.BlockParticipation.Add(stakerID, BlocksMissed{})
}

// GetBlocksMissed return blocks missed num for validator
func (s *Store) GetBlocksMissed(stakerID idx.StakerID) BlocksMissed {
	missedVal, ok := s.cache.BlockParticipation.Get(stakerID)
	if ok {
		if missed, ok := missedVal.(BlocksMissed); ok {
			return missed
		}
	}

	pMissed, _ := s.get(s.table.BlockParticipation, stakerID.Bytes(), &BlocksMissed{}).(*BlocksMissed)
	if pMissed == nil {
		return BlocksMissed{}
	}
	missed := *pMissed

	s.cache.BlockParticipation.Add(stakerID, missed)

	return missed
}

// GetActiveValidationScore return gas value for active validator score
func (s *Store) GetActiveValidationScore(stakerID idx.StakerID) uint64 {
	return s.getValidationScore(s.table.ActiveValidationScore, stakerID)
}

// AddDirtyValidationScore add gas value for active validation score
func (s *Store) AddDirtyValidationScore(stakerID idx.StakerID, gas uint64) {
	s.addValidationScore(s.table.DirtyValidationScore, stakerID, gas)
}

// DelActiveValidationScore deletes record about active validation score of a staker
func (s *Store) DelActiveValidationScore(stakerID idx.StakerID) {
	err := s.table.ActiveValidationScore.Delete(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase key-value", "err", err)
	}
}

// DelDirtyValidationScore deletes record about dirty validation score of a staker
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

// SetValidationScoreCheckpoint set validation score checkpoint time
func (s *Store) SetValidationScoreCheckpoint(cp inter.Timestamp) {
	cpBytes := bigendian.Int64ToBytes(uint64(cp))
	err := s.table.ValidationScoreCheckpoint.Put([]byte(validationScoreCheckpointKey), cpBytes)
	if err != nil {
		s.Log.Crit("Failed to set key-value", "err", err)
	}

	s.cache.ValidationScoreCheckpoint.Add(validationScoreCheckpointKey, cp)
}

// DelAllActiveValidationScores deletes all the record about dirty validation scores of stakers
func (s *Store) DelAllActiveValidationScores() {
	it := s.table.ActiveValidationScore.NewIterator()
	defer it.Release()
	s.dropTable(it, s.table.ActiveValidationScore)
}

// MoveDirtyValidationScoresToActive moves all the dirty records to active
func (s *Store) MoveDirtyValidationScoresToActive() {
	it := s.table.DirtyValidationScore.NewIterator()
	defer it.Release()

	keys := make([][]byte, 0, 500) // don't write during iteration
	vals := make([][]byte, 0, 500)

	for it.Next() {
		keys = append(keys, it.Key())
		vals = append(vals, it.Value())
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

// GetValidationScoreCheckpoint return last validation score checkpoint time
func (s *Store) GetValidationScoreCheckpoint() inter.Timestamp {
	cpVal, ok := s.cache.ValidationScoreCheckpoint.Get(validationScoreCheckpointKey)
	if ok {
		cp, ok := cpVal.(inter.Timestamp)
		if ok {
			return cp
		}
	}

	cpBytes, err := s.table.ValidationScoreCheckpoint.Get([]byte(validationScoreCheckpointKey))
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if cpBytes == nil {
		return 0
	}

	cp := inter.Timestamp(bigendian.BytesToInt64(cpBytes))
	s.cache.ValidationScoreCheckpoint.Add(validationScoreCheckpointKey, cp)

	return cp
}

// GetActiveOriginationScore return gas value for active validator score
func (s *Store) GetActiveOriginationScore(stakerID idx.StakerID) uint64 {
	return s.getOriginationScore(s.table.ActiveOriginationScore, stakerID)
}

// AddDirtyOriginationScore add gas value for active validation score
func (s *Store) AddDirtyOriginationScore(stakerID idx.StakerID, gas uint64) {
	s.addOriginationScore(s.table.DirtyOriginationScore, stakerID, gas)
}

// DelActiveOriginationScore deletes record about active origination score of a staker
func (s *Store) DelActiveOriginationScore(stakerID idx.StakerID) {
	err := s.table.ActiveOriginationScore.Delete(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase key-value", "err", err)
	}
}

// DelDirtyOriginationScore deletes record about dirty origination score of a staker
func (s *Store) DelDirtyOriginationScore(stakerID idx.StakerID) {
	err := s.table.DirtyOriginationScore.Delete(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase key-value", "err", err)
	}
}

// GetDirtyOriginationScore return gas value for active validator score
func (s *Store) GetDirtyOriginationScore(stakerID idx.StakerID) uint64 {
	return s.getOriginationScore(s.table.DirtyOriginationScore, stakerID)
}

func (s *Store) addOriginationScore(t kvdb.KeyValueStore, stakerID idx.StakerID, val uint64) {
	s.mutexes.IncMutex.Lock()
	defer s.mutexes.IncMutex.Unlock()

	score := s.getOriginationScore(t, stakerID)
	score += val
	err := t.Put(stakerID.Bytes(), bigendian.Int64ToBytes(score))
	if err != nil {
		s.Log.Crit("Failed to set key-value", "err", err)
	}
}

func (s *Store) getOriginationScore(t kvdb.KeyValueStore, stakerID idx.StakerID) uint64 {
	scoreBytes, err := t.Get(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if scoreBytes == nil {
		return 0
	}
	return bigendian.BytesToInt64(scoreBytes)
}

// SetOriginationScoreCheckpoint set origination score checkpoint time
func (s *Store) SetOriginationScoreCheckpoint(cp inter.Timestamp) {
	cpBytes := bigendian.Int64ToBytes(uint64(cp))
	err := s.table.OriginationScoreCheckpoint.Put([]byte(originationScoreCheckpointKey), cpBytes)
	if err != nil {
		s.Log.Crit("Failed to set key-value", "err", err)
	}

	s.cache.OriginationScoreCheckpoint.Add(originationScoreCheckpointKey, cp)
}

// DelAllActiveOriginationScores deletes all the record about dirty origination scores of stakers
func (s *Store) DelAllActiveOriginationScores() {
	it := s.table.ActiveOriginationScore.NewIterator()
	defer it.Release()
	s.dropTable(it, s.table.ActiveOriginationScore)
}

// MoveDirtyOriginationScoresToActive moves all the dirty records to active
func (s *Store) MoveDirtyOriginationScoresToActive() {
	it := s.table.DirtyOriginationScore.NewIterator()
	defer it.Release()

	keys := make([][]byte, 0, 500) // don't write during iteration
	vals := make([][]byte, 0, 500)

	for it.Next() {
		keys = append(keys, it.Key())
		vals = append(vals, it.Value())
	}

	for i := range keys {
		err := s.table.ActiveOriginationScore.Put(keys[i], vals[i])
		if err != nil {
			s.Log.Crit("Failed to set key-value", "err", err)
		}
		err = s.table.DirtyOriginationScore.Delete(keys[i])
		if err != nil {
			s.Log.Crit("Failed to erase key-value", "err", err)
		}
	}
}

// GetOriginationScoreCheckpoint return last validation score checkpoint time
func (s *Store) GetOriginationScoreCheckpoint() inter.Timestamp {
	cpVal, ok := s.cache.OriginationScoreCheckpoint.Get(originationScoreCheckpointKey)
	if ok {
		cp, ok := cpVal.(inter.Timestamp)
		if ok {
			return cp
		}
	}

	cpBytes, err := s.table.OriginationScoreCheckpoint.Get([]byte(originationScoreCheckpointKey))
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if cpBytes == nil {
		return 0
	}

	cp := inter.Timestamp(bigendian.BytesToInt64(cpBytes))
	s.cache.OriginationScoreCheckpoint.Add(originationScoreCheckpointKey, cp)

	return cp
}
