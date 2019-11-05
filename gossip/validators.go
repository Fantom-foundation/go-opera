package gossip

import (
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

const (
	PoiPeriodDuration = 2 * 24 * time.Hour
)

// PoiPeriod calculate POI period from int64 unix time
func PoiPeriod(t int64) uint64 {
	return uint64(t / int64(PoiPeriodDuration))
}

// CalcValidatorsPOI calculate and save POI for validator
func (s *Store) CalcValidatorsPOI(stakerID idx.StakerID, delegator common.Address, poiPeriod uint64) {
	vGasUsed := s.GetStakerDelegatorsGasUsed(stakerID)
	dGasUsed := s.GetAddressGasUsed(delegator)

	vGasUsed += dGasUsed
	s.SetStakerDelegatorsGasUsed(stakerID, vGasUsed)

	poi := uint64((vGasUsed * 1000000) / s.GetPOIGasUsed(poiPeriod))
	s.SetValidatorPOI(stakerID, poi)
}

// GetAddressGasUsed get gas used by address
func (s *Store) GetAddressGasUsed(addr common.Address) uint64 {
	gasBytes, err := s.table.AddressGasUsed.Get(addr.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}

	gas := bigendian.BytesToInt64(gasBytes)

	return gas
}

// GetStakerDelegatorsGasUsed get gas used by delegators of a staker
func (s *Store) GetStakerDelegatorsGasUsed(stakerID idx.StakerID) uint64 {
	gasBytes, err := s.table.StakerDelegatorsGasUsed.Get(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}

	gas := bigendian.BytesToInt64(gasBytes)

	return gas
}

// SetStakerDelegatorsGasUsed save gas used by delegators of a staker
func (s *Store) SetStakerDelegatorsGasUsed(stakerID idx.StakerID, gas uint64) {
	gasBytes := bigendian.Int64ToBytes(gas)

	err := s.table.AddressGasUsed.Put(stakerID.Bytes(), gasBytes)
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
}

// SetAddressGasUsed save gas used by address
func (s *Store) SetAddressGasUsed(addr common.Address, gas uint64) {
	gasBytes := bigendian.Int64ToBytes(gas)

	err := s.table.AddressGasUsed.Put(addr.Bytes(), gasBytes)
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
}

// GetAddressLastTxTime get last time for last transaction from this address
func (s *Store) GetAddressLastTxTime(addr common.Address) uint64 {
	gasBytes, err := s.table.AddressLastTxTime.Get(addr.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}

	gas := bigendian.BytesToInt64(gasBytes)

	return gas
}

// SetAddressLastTxTime save last time for trasnaction from this address
func (s *Store) SetAddressLastTxTime(addr common.Address, gas uint64) {
	gasBytes := bigendian.Int64ToBytes(gas)

	err := s.table.AddressLastTxTime.Put(addr.Bytes(), gasBytes)
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
}

// SetPOIGasUsed save gas used for POI period
func (s *Store) SetPOIGasUsed(poiPeriod uint64, gas uint64) {
	key := bigendian.Int64ToBytes(poiPeriod)
	gasBytes := bigendian.Int64ToBytes(gas)

	err := s.table.TotalPOIGasUsed.Put(key, gasBytes)
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
}

// AddPOIGasUsed add gas used to POI period
func (s *Store) AddPOIGasUsed(poiPeriod uint64, gas uint64) {
	oldGas := s.GetPOIGasUsed(poiPeriod)
	s.SetPOIGasUsed(poiPeriod, gas+oldGas)
}

// GetPOIGasUsed get gas used for POI period
func (s *Store) GetPOIGasUsed(poiPeriod uint64) uint64 {
	key := bigendian.Int64ToBytes(poiPeriod)

	gasBytes, err := s.table.TotalPOIGasUsed.Get(key)
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}

	gas := bigendian.BytesToInt64(gasBytes)

	return gas
}

// SetValidatorPOI save POI value for validator address
func (s *Store) SetValidatorPOI(stakerID idx.StakerID, poi uint64) {
	poiBytes := bigendian.Int64ToBytes(poi)
	err := s.table.ValidatorPOIScore.Put(stakerID.Bytes(), poiBytes)
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
}

// GetValidatorPOI get POI value for validator address
func (s *Store) GetValidatorPOI(stakerID idx.StakerID) uint64 {
	poiBytes, err := s.table.ValidatorPOIScore.Get(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}

	poi := bigendian.BytesToInt64(poiBytes)

	return poi
}
