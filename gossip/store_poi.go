package gossip

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// GetAddressGasUsed get gas used by address
func (s *Store) GetAddressGasUsed(addr common.Address, poiPeriod uint64) uint64 {
	key := append(addr.Bytes(), bigendian.Int64ToBytes(poiPeriod)...)
	gasBytes, err := s.table.AddressGasUsed.Get(key)
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}
	if gasBytes == nil {
		return 0
	}

	gas := bigendian.BytesToInt64(gasBytes)

	return gas
}

// SetAddressGasUsed save gas used by address
func (s *Store) SetAddressGasUsed(addr common.Address, poiPeriod uint64, gas uint64) {
	key := append(addr.Bytes(), bigendian.Int64ToBytes(poiPeriod)...)
	gasBytes := bigendian.Int64ToBytes(gas)

	err := s.table.AddressGasUsed.Put(key, gasBytes)
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
}

// GetWeightedDelegatorsGasUsed get gas used by delegators of a staker
func (s *Store) GetWeightedDelegatorsGasUsed(stakerID idx.StakerID) uint64 {
	gasBytes, err := s.table.StakerDelegatorsGasUsed.Get(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}
	if gasBytes == nil {
		return 0
	}

	gas := bigendian.BytesToInt64(gasBytes)

	return gas
}

// SetWeightedDelegatorsGasUsed save gas used by delegators of a staker
func (s *Store) SetWeightedDelegatorsGasUsed(stakerID idx.StakerID, gas uint64) {
	gasBytes := bigendian.Int64ToBytes(gas)

	err := s.table.StakerDelegatorsGasUsed.Put(stakerID.Bytes(), gasBytes)
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
}

func (s *Store) DelWeightedDelegatorsGasUsed(stakerID idx.StakerID) {
	err := s.table.StakerDelegatorsGasUsed.Delete(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase key", "err", err)
	}
}

func (s *Store) DelAllWeightedDelegatorsGasUsed() {
	it := s.table.StakerDelegatorsGasUsed.NewIterator()
	defer it.Release()
	s.dropTable(it, s.table.StakerDelegatorsGasUsed)
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
	if gas == 0 {
		return
	}
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
	if gasBytes == nil {
		return 0
	}

	gas := bigendian.BytesToInt64(gasBytes)

	return gas
}

// SetStakerPOI save POI value for staker
func (s *Store) SetStakerPOI(stakerID idx.StakerID, poi uint64) {
	poiBytes := bigendian.Int64ToBytes(poi)
	err := s.table.StakerPOIScore.Put(stakerID.Bytes(), poiBytes)
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
}

func (s *Store) DelStakerPOI(stakerID idx.StakerID) {
	err := s.table.StakerPOIScore.Delete(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to erase key", "err", err)
	}
}

// GetStakerPOI get POI value for staker
func (s *Store) GetStakerPOI(stakerID idx.StakerID) uint64 {
	poiBytes, err := s.table.StakerPOIScore.Get(stakerID.Bytes())
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
	if poiBytes == nil {
		return 0
	}

	poi := bigendian.BytesToInt64(poiBytes)

	return poi
}

// SetAddressPOI save POI value for a user address
func (s *Store) SetAddressPOI(address common.Address, poi uint64) {
	poiBytes := bigendian.Int64ToBytes(poi)
	err := s.table.AddressPOIScore.Put(address.Bytes(), poiBytes)
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
}

// GetAddressPOI get POI value for user address
func (s *Store) GetAddressPOI(address common.Address) uint64 {
	poiBytes, err := s.table.AddressPOIScore.Get(address.Bytes())
	if err != nil {
		s.Log.Crit("Failed to set key", "err", err)
	}
	if poiBytes == nil {
		return 0
	}

	poi := bigendian.BytesToInt64(poiBytes)

	return poi
}
