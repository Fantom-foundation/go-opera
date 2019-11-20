package gossip

import (
	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
)

func (s *Store) SetEpochValidator(epoch idx.Epoch, stakerID uint64, v *SfcStaker) {
	key := append(epoch.Bytes(), bigendian.Int64ToBytes(stakerID)...)

	s.set(s.table.Validators, key, v)
}

func (s *Store) GetEpochValidator(epoch idx.Epoch, stakerID uint64) *SfcStaker {
	key := append(epoch.Bytes(), bigendian.Int64ToBytes(stakerID)...)

	w, _ := s.get(s.table.Validators, key, &SfcStaker{}).(*SfcStaker)
	return w
}

func (s *Store) SetSfcStaker(stakerID uint64, v *SfcStaker) {
	s.set(s.table.Stakers, bigendian.Int64ToBytes(stakerID), v)

	// Add to LRU cache.
	if s.cache.Stakers != nil {
		s.cache.Stakers.Add(stakerID, v)
	}
}

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

func (s *Store) GetSfcStakers() []SfcStakerAndID {
	it := s.table.Stakers.NewIterator()
	defer it.Release()
	return s.forEachSfcStaker(it)
}

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

		stakerIdBytes := it.Key()[len(it.Key())-8:]
		stakers = append(stakers, SfcStakerAndID{
			StakerID: bigendian.BytesToInt64(stakerIdBytes),
			Staker:   staker,
		})
	}

	return stakers
}

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
