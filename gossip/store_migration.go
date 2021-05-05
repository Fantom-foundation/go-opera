package gossip

import (
	"fmt"

	"github.com/Fantom-foundation/lachesis-base/kvdb"

	"github.com/Fantom-foundation/go-opera/utils/migration"
)

func isEmptyDB(db kvdb.Iteratee) bool {
	it := db.NewIterator(nil, nil)
	defer it.Release()
	return !it.Next()
}

func (s *Store) migrateData() error {
	versions := migration.NewKvdbIDStore(s.table.Version)
	if isEmptyDB(s.mainDB) && isEmptyDB(s.async.mainDB) {
		// short circuit if empty DB
		versions.SetID(s.migrations().ID())
		return nil
	}
	err := s.migrations().Exec(versions)
	if err == nil {
		err = s.Commit()
	}

	return err
}

func (s *Store) migrations() *migration.Migration {
	return migration.
		Begin("opera-gossip-store").
		Next("used gas recovery", s.dataRecovery_UsedGas).
		Next("block's txs recovery", s.dataRecovery_BlocksTxs)
}

func (s *Store) dataRecovery_UsedGas() error {
	start := s.GetGenesisBlockIndex()
	if start == nil {
		return fmt.Errorf("genesis block index is not set")
	}

	for n := *start; true; n++ {
		b := s.GetBlock(n)
		if b == nil {
			break
		}

		var (
			rr                 = s.EvmStore().GetReceipts(n)
			cumulativeGasWrong uint64
			cumulativeGasRight uint64
			fixed              bool
		)
		for i, r := range rr {
			// simulate the bug
			switch {
			case n == *start: // genesis block
				if i == len(b.InternalTxs)-2 || i == len(b.InternalTxs)-1 {
					cumulativeGasWrong = 0
				}
			default: // other blocks
				if i == len(b.InternalTxs)-1 || i == len(b.InternalTxs) {
					cumulativeGasWrong = 0
				}
			}
			// recalc
			gasUsed := r.CumulativeGasUsed - cumulativeGasWrong
			cumulativeGasWrong += gasUsed
			cumulativeGasRight += gasUsed
			// fix
			if r.CumulativeGasUsed != cumulativeGasRight {
				r.CumulativeGasUsed = cumulativeGasRight
				r.GasUsed = gasUsed
				fixed = true
			}
		}
		if fixed {
			s.EvmStore().SetReceipts(n, rr)
		}
	}

	return nil
}

func (s *Store) dataRecovery_BlocksTxs() error {
	start := s.GetGenesisBlockIndex()
	if start == nil {
		return fmt.Errorf("genesis block index is not set")
	}

	for n := *start + 1; true; n++ {
		b := s.GetBlock(n)
		if b == nil {
			break
		}

		var (
			fixed bool
		)

		if fixed {
			s.SetBlock(n, b)
		}
	}

	return nil
}
