package app

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// GetAddressFee get gas used by address
func (s *Store) GetAddressFee(addr common.Address, poiPeriod uint64) *big.Int {
	key := append(addr.Bytes(), bigendian.Int64ToBytes(poiPeriod)...)
	valBytes, err := s.table.AddressFee.Get(key)
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}
	if valBytes == nil {
		return big.NewInt(0)
	}

	val := new(big.Int).SetBytes(valBytes)

	return val
}

// SetAddressFee save gas used by address
func (s *Store) SetAddressFee(addr common.Address, poiPeriod uint64, val *big.Int) {
	key := append(addr.Bytes(), bigendian.Int64ToBytes(poiPeriod)...)

	err := s.table.AddressFee.Put(key, val.Bytes())
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
}

// GetWeightedDelegatorsFee get gas used by delegators of a staker
func (s *Store) GetWeightedDelegatorsFee(stakerID idx.StakerID) *big.Int {
	valBytes, err := s.table.StakerDelegatorsFee.Get(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}
	if valBytes == nil {
		return big.NewInt(0)
	}

	val := new(big.Int).SetBytes(valBytes)

	return val
}

// SetWeightedDelegatorsFee stores gas used by delegators of a staker
func (s *Store) SetWeightedDelegatorsFee(stakerID idx.StakerID, val *big.Int) {
	err := s.table.StakerDelegatorsFee.Put(stakerID.Bytes(), val.Bytes())
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
}

// DelWeightedDelegatorsFee deletes record about gas used by delegators of a staker
func (s *Store) DelWeightedDelegatorsFee(stakerID idx.StakerID) {
	err := s.table.StakerDelegatorsFee.Delete(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase key", "err", err)
	}
}

// DelAllWeightedDelegatorsFee deletes all the records about gas used by delegators of all stakers
func (s *Store) DelAllWeightedDelegatorsFee() {
	it := s.table.StakerDelegatorsFee.NewIterator()
	defer it.Release()
	s.dropTable(it, s.table.StakerDelegatorsFee)
}

// GetAddressLastTxTime get last time for last tx from this address
func (s *Store) GetAddressLastTxTime(addr common.Address) inter.Timestamp {
	tBytes, err := s.table.AddressLastTxTime.Get(addr.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}
	if tBytes == nil {
		return 0
	}

	t := bigendian.BytesToInt64(tBytes)

	return inter.Timestamp(t)
}

// SetAddressLastTxTime save last time for tx from this address
func (s *Store) SetAddressLastTxTime(addr common.Address, t inter.Timestamp) {
	tBytes := bigendian.Int64ToBytes(uint64(t))

	err := s.table.AddressLastTxTime.Put(addr.Bytes(), tBytes)
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
}

// SetPoiFee save gas used for POI period
func (s *Store) SetPoiFee(poiPeriod uint64, val *big.Int) {
	key := bigendian.Int64ToBytes(poiPeriod)

	err := s.table.TotalPoiFee.Put(key, val.Bytes())
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
}

// AddPoiFee add gas used to POI period
func (s *Store) AddPoiFee(poiPeriod uint64, diff *big.Int) {
	if diff.Sign() == 0 {
		return
	}
	val := s.GetPoiFee(poiPeriod)
	val.Add(val, diff)
	s.SetPoiFee(poiPeriod, val)
}

// GetPoiFee get gas used for POI period
func (s *Store) GetPoiFee(poiPeriod uint64) *big.Int {
	key := bigendian.Int64ToBytes(poiPeriod)

	valBytes, err := s.table.TotalPoiFee.Get(key)
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}
	if valBytes == nil {
		return big.NewInt(0)
	}

	val := new(big.Int).SetBytes(valBytes)

	return val
}

// SetStakerPOI save POI value for staker
func (s *Store) SetStakerPOI(stakerID idx.StakerID, poi *big.Int) {
	err := s.table.StakerPOIScore.Put(stakerID.Bytes(), poi.Bytes())
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
}

// DelStakerPOI deletes record about staker's PoI
func (s *Store) DelStakerPOI(stakerID idx.StakerID) {
	err := s.table.StakerPOIScore.Delete(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase key", "err", err)
	}
}

// GetStakerPOI get POI value for staker
func (s *Store) GetStakerPOI(stakerID idx.StakerID) *big.Int {
	poiBytes, err := s.table.StakerPOIScore.Get(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
	if poiBytes == nil {
		return big.NewInt(0)
	}

	poi := new(big.Int).SetBytes(poiBytes)

	return poi
}

// SetAddressPOI save POI value for a user address
func (s *Store) SetAddressPOI(address common.Address, poi *big.Int) {
	err := s.table.AddressPOIScore.Put(address.Bytes(), poi.Bytes())
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
}

// GetAddressPOI get POI value for user address
func (s *Store) GetAddressPOI(address common.Address) *big.Int {
	poiBytes, err := s.table.AddressPOIScore.Get(address.Bytes())
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
	if poiBytes == nil {
		return big.NewInt(0)
	}

	poi := new(big.Int).SetBytes(poiBytes)

	return poi
}
