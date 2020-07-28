package app

import (
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/sfctype"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/pkg/errors"
	"math/big"
)

type legacySfcDelegation struct {
	CreatedEpoch idx.Epoch
	CreatedTime  inter.Timestamp

	DeactivatedEpoch idx.Epoch
	DeactivatedTime  inter.Timestamp

	Amount *big.Int

	ToStakerID idx.StakerID
}

type legacySfcConstants struct {
	ShortGasPowerAllocPerSec uint64
	LongGasPowerAllocPerSec  uint64
	BaseRewardPerSec         *big.Int
}

func (s *Store) MigrateMultiDelegations() error {
	{ // migrate s.table.Delegations
		newKeys := make([][]byte, 0, 10000)
		newValues := make([][]byte, 0, 10000)
		{
			it := s.table.Delegations.NewIterator()
			defer it.Release()
			for it.Next() {
				delegation := &legacySfcDelegation{}
				err := rlp.DecodeBytes(it.Value(), delegation)
				if err != nil {
					return errors.Wrap(err, "failed legacy delegation deserialization during migration")
				}

				addr := common.BytesToAddress(it.Key())
				id := sfctype.DelegationID{
					Delegator: addr,
					StakerID:  delegation.ToStakerID,
				}
				newValue, err := rlp.EncodeToBytes(sfctype.SfcDelegation{
					CreatedEpoch:     delegation.CreatedEpoch,
					CreatedTime:      delegation.CreatedTime,
					DeactivatedEpoch: delegation.DeactivatedEpoch,
					DeactivatedTime:  delegation.DeactivatedTime,
					Amount:           delegation.Amount,
				})
				if err != nil {
					return err
				}

				// don't write into DB during iteration
				newKeys = append(newKeys, id.Bytes())
				newValues = append(newValues, newValue)
			}
		}
		{
			it := s.table.Delegations.NewIterator()
			defer it.Release()
			s.dropTable(it, s.table.Delegations)
		}
		for i := range newKeys {
			err := s.table.Delegations.Put(newKeys[i], newValues[i])
			if err != nil {
				return err
			}
		}
	}
	{ // migrate s.table.DelegationOldRewards
		newKeys := make([][]byte, 0, 10000)
		newValues := make([][]byte, 0, 10000)
		{
			it := s.table.DelegationOldRewards.NewIterator()
			defer it.Release()
			for it.Next() {
				addr := common.BytesToAddress(it.Key())
				delegations := s.GetSfcDelegationsByAddr(addr, 2)
				if len(delegations) > 1 {
					return errors.New("more than one delegation during multi-delegation migration")
				}
				if len(delegations) == 0 {
					continue
				}
				toStakerID := delegations[0].ID.StakerID
				id := sfctype.DelegationID{
					Delegator: addr,
					StakerID:  toStakerID,
				}

				// don't write into DB during iteration
				newKeys = append(newKeys, id.Bytes())
				newValues = append(newKeys, it.Value())
			}
		}
		{
			it := s.table.DelegationOldRewards.NewIterator()
			defer it.Release()
			s.dropTable(it, s.table.DelegationOldRewards)
		}
		for i := range newKeys {
			err := s.table.DelegationOldRewards.Put(newKeys[i], newValues[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) MigrateEraseGenesisField() error {
	it := s.mainDb.NewIteratorWithPrefix([]byte("G"))
	defer it.Release()
	s.dropTable(it, s.mainDb)
	return nil
}

func (s *Store) MigrateAdjustableOfflinePeriod() error {
	{ // migrate s.table.SfcConstants
		newKeys := make([][]byte, 0, 10000)
		newValues := make([][]byte, 0, 10000)
		{
			it := s.table.SfcConstants.NewIterator()
			defer it.Release()
			for it.Next() {
				constants := &legacySfcConstants{}
				err := rlp.DecodeBytes(it.Value(), constants)
				if err != nil {
					return errors.Wrap(err, "failed legacy constants deserialization during migration")
				}

				newConstants := SfcConstants{
					ShortGasPowerAllocPerSec: constants.ShortGasPowerAllocPerSec,
					LongGasPowerAllocPerSec:  constants.LongGasPowerAllocPerSec,
					BaseRewardPerSec:         constants.BaseRewardPerSec,
				}
				newValue, err := rlp.EncodeToBytes(newConstants)
				if err != nil {
					return err
				}

				// don't write into DB during iteration
				newKeys = append(newKeys, it.Key())
				newValues = append(newValues, newValue)
			}
		}
		{
			it := s.table.SfcConstants.NewIterator()
			defer it.Release()
			s.dropTable(it, s.table.SfcConstants)
		}
		for i := range newKeys {
			err := s.table.SfcConstants.Put(newKeys[i], newValues[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}
