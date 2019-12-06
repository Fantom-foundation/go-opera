package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
)

// SetEpochValidator stores EpochValidator
func (s *Store) SetEpochValidator(epoch idx.Epoch, stakerID uint64, v *SfcStaker) {
	key := append(epoch.Bytes(), bigendian.Int64ToBytes(stakerID)...)

	s.set(s.table.Validators, key, v)
}

// GetEpochValidator returns stored EpochValidator
func (s *Store) GetEpochValidator(epoch idx.Epoch, stakerID uint64) *SfcStaker {
	key := append(epoch.Bytes(), bigendian.Int64ToBytes(stakerID)...)

	w, _ := s.get(s.table.Validators, key, &SfcStaker{}).(*SfcStaker)
	return w
}

// SetSfcStaker stores SfcStaker
func (s *Store) SetSfcStaker(stakerID uint64, v *SfcStaker) {
	s.set(s.table.Stakers, bigendian.Int64ToBytes(stakerID), v)

	// Add to LRU cache.
	if s.cache.Stakers != nil {
		s.cache.Stakers.Add(stakerID, v)
	}
}

// DelSfcStaker deletes SfcStaker
func (s *Store) DelSfcStaker(stakerID uint64) {
	err := s.table.Stakers.Delete(bigendian.Int64ToBytes(stakerID))
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
	return s.forEachSfcStaker(it)
}

// GetEpochValidators returns all stored EpochValidators on the epoch
func (s *Store) GetEpochValidators(epoch idx.Epoch) []SfcStakerAndID {
	it := s.table.Validators.NewIteratorWithPrefix(epoch.Bytes())
	defer it.Release()
	return s.forEachSfcStaker(it)
}

func (s *Store) forEachSfcStaker(it ethdb.Iterator) []SfcStakerAndID {
	stakers := make([]SfcStakerAndID, 0, 1000)
	for it.Next() {
		staker := &SfcStaker{}
		err := rlp.DecodeBytes(it.Value(), staker)
		if err != nil {
			s.Log.Crit("Failed to decode rlp", "err", err)
		}

		stakerIDBytes := it.Key()[len(it.Key())-8:]
		stakers = append(stakers, SfcStakerAndID{
			StakerID: bigendian.BytesToInt64(stakerIDBytes),
			Staker:   staker,
		})
	}

	return stakers
}

// GetSfcStaker returns stored SfcStaker
func (s *Store) GetSfcStaker(stakerID uint64) *SfcStaker {
	// Get data from LRU cache first.
	if s.cache.Stakers != nil {
		if c, ok := s.cache.Stakers.Get(stakerID); ok {
			if b, ok := c.(*SfcStaker); ok {
				return b
			}
		}
	}

	w, _ := s.get(s.table.Stakers, bigendian.Int64ToBytes(stakerID), &SfcStaker{}).(*SfcStaker)

	// Add to LRU cache.
	if w != nil && s.cache.Stakers != nil {
		s.cache.Stakers.Add(stakerID, w)
	}

	return w
}
