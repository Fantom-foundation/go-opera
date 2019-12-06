package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
)

// SetEpochValidator stores EpochValidator
func (s *Store) SetEpochValidator(epoch idx.Epoch, stakerID idx.StakerID, v *SfcStaker) {
	key := append(epoch.Bytes(), stakerID.Bytes()...)

	s.set(s.table.Validators, key, v)
}

// GetEpochValidator returns stored EpochValidator
func (s *Store) GetEpochValidator(epoch idx.Epoch, stakerID idx.StakerID) *SfcStaker {
	key := append(epoch.Bytes(), stakerID.Bytes()...)

	w, _ := s.get(s.table.Validators, key, &SfcStaker{}).(*SfcStaker)
	return w
}

// DelSfcStaker deletes SfcStaker
func (s *Store) SetSfcStaker(stakerID idx.StakerID, v *SfcStaker) {
	s.set(s.table.Stakers, stakerID.Bytes(), v)

	// Add to LRU cache.
	if s.cache.Stakers != nil {
		s.cache.Stakers.Add(stakerID, v)
	}
}

// DelSfcStaker deletes SfcStaker
func (s *Store) DelSfcStaker(stakerID idx.StakerID) {
	err := s.table.Stakers.Delete(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase staker")
	}

	// Add to LRU cache.
	if s.cache.Stakers != nil {
		s.cache.Stakers.Remove(stakerID)
	}
}

// GetSfcStakers returns all stored SfcStakers
func (s *Store) GetSfcStakers() []SfcStakerAndID {
	it := s.table.Stakers.NewIterator()
	defer it.Release()
	return s.getSfcStakers(it)
}

// GetEpochValidators returns all stored EpochValidators on the epoch
func (s *Store) GetEpochValidators(epoch idx.Epoch) []SfcStakerAndID {
	it := s.table.Validators.NewIteratorWithPrefix(epoch.Bytes())
	defer it.Release()
	return s.getSfcStakers(it)
}

func (s *Store) getSfcStakers(it ethdb.Iterator) []SfcStakerAndID {
	stakers := make([]SfcStakerAndID, 0, 1000)
	for it.Next() {
		staker := &SfcStaker{}
		err := rlp.DecodeBytes(it.Value(), staker)
		if err != nil {
			s.Log.Crit("Failed to decode rlp while iteration", "err", err)
		}

		stakerIDBytes := it.Key()[len(it.Key())-4:]
		stakers = append(stakers, SfcStakerAndID{
			StakerID: idx.BytesToStakerID(stakerIDBytes),
			Staker:   staker,
		})
	}

	return stakers
}

// GetSfcStaker returns stored SfcStaker
func (s *Store) GetSfcStaker(stakerID idx.StakerID) *SfcStaker {
	// Get data from LRU cache first.
	if s.cache.Stakers != nil {
		if c, ok := s.cache.Stakers.Get(stakerID); ok {
			if b, ok := c.(*SfcStaker); ok {
				return b
			}
		}
	}

	w, _ := s.get(s.table.Stakers, stakerID.Bytes(), &SfcStaker{}).(*SfcStaker)

	// Add to LRU cache.
	if w != nil && s.cache.Stakers != nil {
		s.cache.Stakers.Add(stakerID, w)
	}

	return w
}
