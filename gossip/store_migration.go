package gossip

import (
	"fmt"

	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/common"

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
		Next("fix-of-used-gas", s.fixOfUsedGas)
}

func (s *Store) fixOfUsedGas() error {
	start := s.GetGenesisBlockIndex()
	if start == nil {
		return fmt.Errorf("genesis block index is not set")
	}

	var (
		minBlock  idx.Block
		maxBlock  idx.Block
		fixedTxs  uint
		badBlocks uint // TODO: fix blocks data
	)

	for n := *start; true; n++ {
		b := s.GetBlock(n)
		if b == nil {
			break
		}

		var (
			cumulativeGasWrong uint64
			cumulativeGasRight uint64
		)
		rr := s.EvmStore().GetReceipts(n)

		txs := b.NotSkippedTxs()
		if len(txs) != len(rr) {
			badBlocks++
			txs = make([]common.Hash, len(rr))
		}

		for i, r := range rr {
			// simulate the bug
			if i == len(b.InternalTxs)-1 || i == len(b.InternalTxs) {
				cumulativeGasWrong = 0
			}
			// restore
			gasUsed := r.CumulativeGasUsed - cumulativeGasWrong
			cumulativeGasWrong += gasUsed
			cumulativeGasRight += gasUsed

			if r.CumulativeGasUsed != cumulativeGasWrong {
				panic(fmt.Sprintf(
					"B %d[%d] %s : %d(%d) != %d(%d)\n", n, i, txs[i].Hex(), r.GasUsed, r.CumulativeGasUsed, gasUsed, cumulativeGasWrong))
			}

			if r.CumulativeGasUsed != cumulativeGasRight {
				fmt.Printf(
					"B %d[%d] %s : %d(%d) --> %d(%d)\n", n, i, txs[i].Hex(), r.GasUsed, r.CumulativeGasUsed, gasUsed, cumulativeGasRight)

				if minBlock == 0 {
					minBlock = n
				}
				maxBlock = n
				fixedTxs++
			}
		}
	}

	panic(fmt.Sprintf("Start %d. Finish %d-%d (%d), badBlocks %d\n", *start, minBlock, maxBlock, fixedTxs, badBlocks))
	return nil
}
