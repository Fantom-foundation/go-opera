package app

import (
	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
)

// SetGasPowerRefund stores amount of gas power to refund
func (s *Store) SetGasPowerRefund(epoch idx.Epoch, stakerID idx.StakerID, refund uint64) {
	key := append(epoch.Bytes(), stakerID.Bytes()...)

	err := s.table.GasPowerRefund.Put(key, bigendian.Int64ToBytes(refund))
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// GetGasPowerRefund returns stored amount of gas power to refund
func (s *Store) GetGasPowerRefund(epoch idx.Epoch, stakerID idx.StakerID) uint64 {
	key := append(epoch.Bytes(), stakerID.Bytes()...)

	refundBytes, err := s.table.GasPowerRefund.Get(key)
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
	if refundBytes == nil {
		return 0
	}
	return bigendian.BytesToInt64(refundBytes)
}

// GetGasPowerRefunds returns all stored amount of gas power to refund
func (s *Store) GetGasPowerRefunds(epoch idx.Epoch) map[idx.StakerID]uint64 {
	rr := make(map[idx.StakerID]uint64)

	it := s.table.GasPowerRefund.NewIteratorWithPrefix(epoch.Bytes())
	defer it.Release()
	for it.Next() {
		creator := idx.BytesToStakerID(it.Key()[4:])
		gasPower := bigendian.BytesToInt64(it.Value())
		rr[creator] = gasPower
	}

	return rr
}

// IncGasPowerRefund increments amount of gas power to refund
func (s *Store) IncGasPowerRefund(epoch idx.Epoch, stakerID idx.StakerID, diff uint64) {
	if diff == 0 {
		return
	}
	refund := s.GetGasPowerRefund(epoch, stakerID)
	refund += diff
	s.SetGasPowerRefund(epoch, stakerID, refund)
}

// DelGasPowerRefunds erases all record on epoch
func (s *Store) DelGasPowerRefunds(epoch idx.Epoch) {
	it := s.table.GasPowerRefund.NewIteratorWithPrefix(epoch.Bytes())
	defer it.Release()
	s.dropTable(it, s.table.GasPowerRefund)
}
