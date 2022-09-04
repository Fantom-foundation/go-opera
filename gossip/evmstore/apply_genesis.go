package evmstore

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/batched"

	"github.com/Fantom-foundation/go-opera/opera/genesis"
)

// ApplyGenesis writes initial state.
func (s *Store) ApplyGenesis(g genesis.Genesis) (err error) {
	batch := s.EvmDb.NewBatch()
	defer batch.Reset()
	g.RawEvmItems.ForEach(func(key, value []byte) bool {
		if err != nil {
			return false
		}
		err = batch.Put(key, value)
		if err != nil {
			return false
		}
		if batch.ValueSize() > kvdb.IdealBatchSize {
			err = batch.Write()
			if err != nil {
				return false
			}
			batch.Reset()
		}
		return true
	})
	if err != nil {
		return err
	}
	return batch.Write()
}

func (s *Store) WrapTablesAsBatched() (unwrap func()) {
	origTables := s.table

	batchedTxs := batched.Wrap(s.table.Txs)
	s.table.Txs = batchedTxs

	batchedTxPositions := batched.Wrap(s.table.TxPositions)
	s.table.TxPositions = batchedTxPositions

	unwrapLogs := s.EvmLogs.WrapTablesAsBatched()

	batchedReceipts := batched.Wrap(s.table.Receipts)
	s.table.Receipts = batchedReceipts
	return func() {
		_ = batchedTxs.Flush()
		_ = batchedTxPositions.Flush()
		_ = batchedReceipts.Flush()
		unwrapLogs()
		s.table = origTables
	}
}
