package gossip

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/idx"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/inter"
	"github.com/Fantom-foundation/go-opera/inter/iblockproc"
	"github.com/Fantom-foundation/go-opera/utils/migration"
)

func isEmptyDB(db kvdb.Iteratee) bool {
	it := db.NewIterator(nil, nil)
	defer it.Release()
	return !it.Next()
}

func (s *Store) migrateData() error {
	versions := migration.NewKvdbIDStore(s.table.Version)
	if isEmptyDB(s.table.Version) {
		// short circuit if empty DB
		versions.SetID(s.migrations().ID())
		return nil
	}

	err := s.migrations().Exec(versions, s.flushDBs)
	if err != nil {
		panic(err)
	}
	return err
}

func (s *Store) migrations() *migration.Migration {
	return migration.
		Begin("opera-gossip-store").
		Next("used gas recovery", unsupportedMigration).
		Next("tx hashes recovery", unsupportedMigration).
		Next("DAG heads recovery", unsupportedMigration).
		Next("DAG last events recovery", unsupportedMigration).
		Next("BlockState recovery", unsupportedMigration).
		Next("LlrState recovery", s.recoverLlrState).
		Next("erase gossip-async db", s.eraseGossipAsyncDB).
		Next("erase SFC API table", s.eraseSfcApiTable).
		Next("erase legacy genesis DB", s.eraseGenesisDB).
		Next("calculate upgrade heights", s.calculateUpgradeHeights).
		Next("EVM TxPosition.BlockOffset fix", s.fixTxPositionBlockOffset)
}

func unsupportedMigration() error {
	return fmt.Errorf("DB version isn't supported, please restart from scratch")
}

var (
	fixTxHash1  = common.HexToHash("0xb6840d4c0eb562b0b1731760223d91b36edc6c160958e23e773e6058eea30458")
	fixTxEvent1 = hash.HexToEventHash("0x00001718000003d4d3955bf592e12fb80a60574fa4b18bd5805b4c010d75e86d")
	fixTxHash2  = common.HexToHash("0x3aeede91740093cb8feb1296a34cf70d86d2f802cff860edd798978e94a40534")
	fixTxEvent2 = hash.HexToEventHash("0x0000179e00000c464d756a7614d0ca067fcb37ee4452004bf308c9df561e85e8")
)

const (
	fixTxEventPos1 = 2
	fixTxBlock1    = 4738821
	fixTxEventPos2 = 0
	fixTxBlock2    = 4801307
)

func fixEventTxHashes(e *inter.EventPayload) {
	if e.ID() == fixTxEvent1 {
		e.Txs()[fixTxEventPos1].SetHash(fixTxHash1)
	}
	if e.ID() == fixTxEvent2 {
		e.Txs()[fixTxEventPos2].SetHash(fixTxHash2)
	}
}

func (s *Store) recoverLlrState() error {
	v1, ok := s.rlp.Get(s.table.BlockEpochState, []byte(sKey), &BlockEpochState{}).(*BlockEpochState)
	if !ok {
		return errors.New("epoch state reading failed: genesis not applied")
	}

	epoch := v1.EpochState.Epoch + 1
	block := v1.BlockState.LastBlock.Idx + 1

	s.setLlrState(LlrState{
		LowestEpochToDecide: epoch,
		LowestEpochToFill:   epoch,
		LowestBlockToDecide: block,
		LowestBlockToFill:   block,
	})
	s.FlushLlrState()
	return nil
}

func (s *Store) eraseSfcApiTable() error {
	sfcapiTable, _ := s.dbs.OpenDB("gossip/S")
	it := sfcapiTable.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		err := sfcapiTable.Delete(it.Key())
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) eraseGossipAsyncDB() error {
	asyncDB, err := s.dbs.OpenDB("gossip-async")
	if err != nil {
		return fmt.Errorf("failed to open gossip-async to drop: %v", err)
	}

	_ = asyncDB.Close()
	asyncDB.Drop()

	return nil
}

func (s *Store) eraseGenesisDB() error {
	genesisDB, err := s.dbs.OpenDB("genesis")
	if err != nil {
		return nil
	}

	_ = genesisDB.Close()
	genesisDB.Drop()
	return nil
}

func (s *Store) calculateUpgradeHeights() error {
	var prevEs *iblockproc.EpochState
	s.ForEachHistoryBlockEpochState(func(bs iblockproc.BlockState, es iblockproc.EpochState) bool {
		s.WriteUpgradeHeight(bs, es, prevEs)
		prevEs = &es
		return true
	})
	if prevEs == nil {
		// special case when no history is available
		s.WriteUpgradeHeight(s.GetBlockState(), s.GetEpochState(), nil)
	}
	return nil
}

func (s *Store) fixTxPositionBlockOffset() (err error) {
	const parallels = 10
	var (
		wg    sync.WaitGroup
		items = new(uint32)
	)

	receiptsTable, _ := s.dbs.OpenDB("evm/r")
	txPositionsTable, _ := s.dbs.OpenDB("evm/x")

	processBlockReceipts := func(input <-chan []*types.ReceiptForStorage) {
		defer wg.Done()
		pos := new(evmstore.TxPosition)
		for rr := range input {
			for _, r := range rr {
				key := r.TxHash.Bytes()
				got := s.rlp.Get(txPositionsTable, key, pos)
				if got == nil {
					continue
				}
				pos.BlockOffset = uint32(r.TransactionIndex)
				s.rlp.Set(txPositionsTable, key, pos)

				atomic.AddUint32(items, 1)
			}
		}
	}

	wg.Add(parallels)
	threads := make([]chan []*types.ReceiptForStorage, parallels)
	for i := range threads {
		threads[i] = make(chan []*types.ReceiptForStorage, 10)
		go processBlockReceipts(threads[i])
	}

	var (
		block         idx.Block
		start         = time.Now()
		prevFlushTime = time.Now()
	)
	it := receiptsTable.NewIterator(nil, nil)
	defer it.Release()
	for n := 0; it.Next(); n++ {
		block = idx.BytesToBlock(it.Key())

		var receiptsStorage []*types.ReceiptForStorage
		err := rlp.DecodeBytes(it.Value(), &receiptsStorage)
		if err != nil {
			s.Log.Crit("Failed to decode rlp", "err", err, "size", len(it.Value()))
		}
		threads[n%parallels] <- receiptsStorage

		if s.dbs.NotFlushedSizeEst() > s.cfg.MaxNonFlushedSize/2 || time.Since(prevFlushTime) > s.cfg.MaxNonFlushedPeriod {
			prevFlushTime = time.Now()
			err = s.flushDBs()
			if err != nil {
				break
			}
			s.Log.Info("Txs positions processing", "elapsed", common.PrettyDuration(time.Since(start)), "block", block, "items", atomic.LoadUint32(items))
		}
	}
	for i := range threads {
		close(threads[i])
	}
	wg.Wait()
	// no need to flush dbs at end as it migration engine does

	s.Log.Info("Txs positions processing", "elapsed", common.PrettyDuration(time.Since(start)), "block", block, "items", *items)

	return
}
