package app

import (
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/sfctype"
)

// SetEpochValidator stores EpochValidator
func (s *Store) SetEpochValidators(epoch idx.Epoch, vv []sfctype.SfcStakerAndID) {
	for _, v := range vv {
		key := append(epoch.Bytes(), v.StakerID.Bytes()...)
		s.set(s.table.Validators, key, v.Staker)
	}

	// Add to LRU cache.
	s.cache.Validators.Add(epoch, vv)
}

// HasEpochValidator returns true if validator exists
func (s *Store) HasEpochValidator(epoch idx.Epoch, stakerID idx.StakerID) bool {
	key := append(epoch.Bytes(), stakerID.Bytes()...)
	return s.has(s.table.Validators, key)
}

// SetSfcStaker stores SfcStaker
func (s *Store) SetSfcStaker(stakerID idx.StakerID, v *sfctype.SfcStaker) {
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

// ForEachSfcStaker iterates all stored SfcStakers
func (s *Store) ForEachSfcStaker(do func(sfctype.SfcStakerAndID)) {
	it := s.table.Stakers.NewIterator()
	defer it.Release()
	s.forEachSfcStaker(it, do)
}

// GetSfcStakers returns all stored SfcStakers
func (s *Store) GetSfcStakers() []sfctype.SfcStakerAndID {
	stakers := make([]sfctype.SfcStakerAndID, 0, 200)
	s.ForEachSfcStaker(func(it sfctype.SfcStakerAndID) {
		stakers = append(stakers, it)
	})
	return stakers
}

// GetEpochValidators returns all stored EpochValidators on the epoch
func (s *Store) GetEpochValidators(epoch idx.Epoch) []sfctype.SfcStakerAndID {
	// Get from cache
	if bVal, okGet := s.cache.Validators.Get(epoch); okGet {
		return bVal.([]sfctype.SfcStakerAndID)
	}

	it := s.table.Validators.NewIteratorWithPrefix(epoch.Bytes())
	defer it.Release()
	validators := make([]sfctype.SfcStakerAndID, 0, 200)
	s.forEachSfcStaker(it, func(staker sfctype.SfcStakerAndID) {
		validators = append(validators, staker)
	})

	// Add to LRU cache.
	s.cache.Validators.Add(epoch, validators)

	return validators
}

func (s *Store) forEachSfcStaker(it ethdb.Iterator, do func(sfctype.SfcStakerAndID)) {
	for it.Next() {
		staker := &sfctype.SfcStaker{}
		err := rlp.DecodeBytes(it.Value(), staker)
		if err != nil {
			s.Log.Crit("Failed to decode rlp while iteration", "err", err)
		}

		stakerIDBytes := it.Key()[len(it.Key())-4:]
		do(sfctype.SfcStakerAndID{
			StakerID: idx.BytesToStakerID(stakerIDBytes),
			Staker:   staker,
		})
	}
}

// GetSfcStaker returns stored SfcStaker
func (s *Store) GetSfcStaker(stakerID idx.StakerID) *sfctype.SfcStaker {
	// Get data from LRU cache first.
	if s.cache.Stakers != nil {
		if c, ok := s.cache.Stakers.Get(stakerID); ok {
			if b, ok := c.(*sfctype.SfcStaker); ok {
				return b
			}
		}
	}

	w, _ := s.get(s.table.Stakers, stakerID.Bytes(), &sfctype.SfcStaker{}).(*sfctype.SfcStaker)

	// Add to LRU cache.
	if w != nil && s.cache.Stakers != nil {
		s.cache.Stakers.Add(stakerID, w)
	}

	return w
}

// HasSfcStaker returns true if staker exists
func (s *Store) HasSfcStaker(stakerID idx.StakerID) bool {
	ok, err := s.table.Stakers.Has(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get staker", "err", err)
	}
	return ok
}
