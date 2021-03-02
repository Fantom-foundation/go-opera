package sfcapi

import (
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
)

// SetEpochValidators stores EpochValidators
func (s *Store) SetEpochValidators(epoch idx.Epoch, vv []SfcStakerAndID) {
	for _, v := range vv {
		key := append(epoch.Bytes(), v.StakerID.Bytes()...)
		s.rlp.Set(s.table.Validators, key, v.Staker)
	}
}

// HasEpochValidator returns true if validator exists
func (s *Store) HasEpochValidator(epoch idx.Epoch, stakerID idx.ValidatorID) bool {
	key := append(epoch.Bytes(), stakerID.Bytes()...)
	ok, _ := s.table.Validators.Has(key)
	return ok
}

// SetSfcStaker stores SfcStaker
func (s *Store) SetSfcStaker(stakerID idx.ValidatorID, v *SfcStaker) {
	s.rlp.Set(s.table.Stakers, stakerID.Bytes(), v)
}

// DelSfcStaker deletes SfcStaker
func (s *Store) DelSfcStaker(stakerID idx.ValidatorID) {
	err := s.table.Stakers.Delete(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase staker")
	}
}

// ForEachSfcStaker iterates all stored SfcStakers
func (s *Store) ForEachSfcStaker(do func(SfcStakerAndID)) {
	it := s.table.Stakers.NewIterator(nil, nil)
	defer it.Release()
	s.forEachSfcStaker(it, do)
}

// GetSfcStakers returns all stored SfcStakers
func (s *Store) GetSfcStakers() []SfcStakerAndID {
	stakers := make([]SfcStakerAndID, 0, 200)
	s.ForEachSfcStaker(func(it SfcStakerAndID) {
		stakers = append(stakers, it)
	})
	return stakers
}

// GetEpochValidators returns all stored EpochValidators on the epoch
func (s *Store) GetEpochValidators(epoch idx.Epoch) []SfcStakerAndID {
	it := s.table.Validators.NewIterator(epoch.Bytes(), nil)
	defer it.Release()
	validators := make([]SfcStakerAndID, 0, 200)
	s.forEachSfcStaker(it, func(staker SfcStakerAndID) {
		validators = append(validators, staker)
	})

	return validators
}

func (s *Store) forEachSfcStaker(it ethdb.Iterator, do func(SfcStakerAndID)) {
	for it.Next() {
		staker := &SfcStaker{}
		err := rlp.DecodeBytes(it.Value(), staker)
		if err != nil {
			s.Log.Crit("Failed to decode rlp while iteration", "err", err)
		}

		stakerIDBytes := it.Key()[len(it.Key())-4:]
		do(SfcStakerAndID{
			StakerID: idx.BytesToValidatorID(stakerIDBytes),
			Staker:   staker,
		})
	}
}

// GetSfcStaker returns stored SfcStaker
func (s *Store) GetSfcStaker(stakerID idx.ValidatorID) *SfcStaker {
	w, _ := s.rlp.Get(s.table.Stakers, stakerID.Bytes(), &SfcStaker{}).(*SfcStaker)

	return w
}

// HasSfcStaker returns true if staker exists
func (s *Store) HasSfcStaker(stakerID idx.ValidatorID) bool {
	ok, err := s.table.Stakers.Has(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get staker", "err", err)
	}
	return ok
}
