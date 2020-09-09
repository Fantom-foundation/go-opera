package app

import (
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"

	"github.com/Fantom-foundation/go-opera/inter/sfctype"
)

// SetSfcStaker stores SfcStaker
func (s *Store) SetSfcStaker(validatorID idx.ValidatorID, v *sfctype.SfcStaker) {
	s.set(s.table.Stakers, validatorID.Bytes(), v)

	// Add to LRU cache.
	if s.cache.Stakers != nil {
		s.cache.Stakers.Add(validatorID, v)
	}
}

// DelSfcStaker deletes SfcStaker
func (s *Store) DelSfcStaker(validatorID idx.ValidatorID) {
	err := s.table.Stakers.Delete(validatorID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase staker")
	}

	// Add to LRU cache.
	if s.cache.Stakers != nil {
		s.cache.Stakers.Remove(validatorID)
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

func (s *Store) forEachSfcStaker(it ethdb.Iterator, do func(sfctype.SfcStakerAndID)) {
	for it.Next() {
		staker := &sfctype.SfcStaker{}
		err := rlp.DecodeBytes(it.Value(), staker)
		if err != nil {
			s.Log.Crit("Failed to decode rlp while iteration", "err", err)
		}

		validatorIDBytes := it.Key()[len(it.Key())-4:]
		do(sfctype.SfcStakerAndID{
			ValidatorID: idx.BytesToValidatorID(validatorIDBytes),
			Staker:      staker,
		})
	}
}

// GetSfcStaker returns stored SfcStaker
func (s *Store) GetSfcStaker(validatorID idx.ValidatorID) *sfctype.SfcStaker {
	// Get data from LRU cache first.
	if s.cache.Stakers != nil {
		if c, ok := s.cache.Stakers.Get(validatorID); ok {
			if b, ok := c.(*sfctype.SfcStaker); ok {
				return b
			}
		}
	}

	w, _ := s.get(s.table.Stakers, validatorID.Bytes(), &sfctype.SfcStaker{}).(*sfctype.SfcStaker)

	// Add to LRU cache.
	if w != nil && s.cache.Stakers != nil {
		s.cache.Stakers.Add(validatorID, w)
	}

	return w
}

// HasSfcStaker returns true if staker exists
func (s *Store) HasSfcStaker(validatorID idx.ValidatorID) bool {
	ok, err := s.table.Stakers.Has(validatorID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get staker", "err", err)
	}
	return ok
}
