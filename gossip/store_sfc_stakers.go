package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/sfctype"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
)

// SetEpochValidator stores EpochValidator
func (s *Store) SetEpochValidator(epoch idx.Epoch, stakerID idx.StakerID, v *sfctype.SfcStaker) {
	key := append(epoch.Bytes(), stakerID.Bytes()...)

	s.set(s.table.Validators, key, v)
}

// GetEpochValidator returns stored EpochValidator
func (s *Store) GetEpochValidator(epoch idx.Epoch, stakerID idx.StakerID) *sfctype.SfcStaker {
	key := append(epoch.Bytes(), stakerID.Bytes()...)

	w, _ := s.get(s.table.Validators, key, &sfctype.SfcStaker{}).(*sfctype.SfcStaker)
	return w
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

// GetEpochValidators returns all stored EpochValidators on the epoch
func (s *Store) GetEpochValidators(epoch idx.Epoch) []sfctype.SfcStakerAndID {
	it := s.table.Validators.NewIteratorWithPrefix(epoch.Bytes())
	defer it.Release()
	validators := make([]sfctype.SfcStakerAndID, 0, 200)
	s.forEachSfcStaker(it, func(staker sfctype.SfcStakerAndID) {
		validators = append(validators, staker)
	})
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

// HasEpochValidator returns true if validator exists
func (s *Store) HasEpochValidator(epoch idx.Epoch, stakerID idx.StakerID) bool {
	key := append(epoch.Bytes(), stakerID.Bytes()...)

	ok, err := s.table.Validators.Has(key)
	if err != nil {
		s.Log.Crit("Failed to get staker", "err", err)
	}
	return ok
}
