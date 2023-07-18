package gossip

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/flushable"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/Fantom-foundation/lachesis-base/utils/wlru"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/gossip/txtrace"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/utils/adapters/snap2kvdb"
	"github.com/Fantom-foundation/go-opera/utils/dbutil/switchable"
	"github.com/Fantom-foundation/go-opera/utils/eventid"
	"github.com/Fantom-foundation/go-opera/utils/randat"
	"github.com/Fantom-foundation/go-opera/utils/rlpstore"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	dbs kvdb.FlushableDBProducer
	cfg StoreConfig

	snapshotedEVMDB *switchable.Snapshot
	evm             *evmstore.Store
	txtrace         *txtrace.Store
	table           struct {
		Version kvdb.Store `table:"_"`

		// Main DAG tables
		BlockEpochState        kvdb.Store `table:"D"`
		BlockEpochStateHistory kvdb.Store `table:"h"`
		Events                 kvdb.Store `table:"e"`
		Blocks                 kvdb.Store `table:"b"`
		EpochBlocks            kvdb.Store `table:"P"`
		Genesis                kvdb.Store `table:"g"`
		UpgradeHeights         kvdb.Store `table:"U"`

		// Transaction traces
		TransactionTraces kvdb.Store `table:"t"`

		// P2P-only
		HighestLamport kvdb.Store `table:"l"`

		// Network version
		NetworkVersion kvdb.Store `table:"V"`

		// API-only
		BlockHashes kvdb.Store `table:"B"`

		LlrState           kvdb.Store `table:"S"`
		LlrBlockResults    kvdb.Store `table:"R"`
		LlrEpochResults    kvdb.Store `table:"Q"`
		LlrBlockVotes      kvdb.Store `table:"T"`
		LlrBlockVotesIndex kvdb.Store `table:"J"`
		LlrEpochVotes      kvdb.Store `table:"E"`
		LlrEpochVoteIndex  kvdb.Store `table:"I"`
		LlrLastBlockVotes  kvdb.Store `table:"G"`
		LlrLastEpochVote   kvdb.Store `table:"F"`
	}

	prevFlushTime time.Time

	epochStore atomic.Value

	cache struct {
		Events                 *wlru.Cache `cache:"-"` // store by pointer
		EventIDs               *eventid.Cache
		EventsHeaders          *wlru.Cache  `cache:"-"` // store by pointer
		Blocks                 *wlru.Cache  `cache:"-"` // store by pointer
		BlockHashes            *wlru.Cache  `cache:"-"` // store by value
		BRHashes               *wlru.Cache  `cache:"-"` // store by value
		EvmBlocks              *wlru.Cache  `cache:"-"` // store by pointer
		BlockEpochStateHistory *wlru.Cache  `cache:"-"` // store by pointer
		BlockEpochState        atomic.Value // store by value
		HighestLamport         atomic.Value // store by value
		LastBVs                atomic.Value // store by pointer
		LastEV                 atomic.Value // store by pointer
		LlrState               atomic.Value // store by value
		KvdbEvmSnap            atomic.Value // store by pointer
		UpgradeHeights         atomic.Value // store by pointer
		Genesis                atomic.Value // store by value
		LlrBlockVotesIndex     *VotesCache  // store by pointer
		LlrEpochVoteIndex      *VotesCache  // store by pointer
	}

	mutex struct {
		WriteLlrState sync.Mutex
	}

	rlp rlpstore.Helper

	logger.Instance
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	mems := memorydb.NewProducer("")
	dbs := flushable.NewSyncedPool(mems, []byte{0})
	cfg := LiteStoreConfig()

	return NewStore(dbs, cfg)
}

// NewStore creates store over key-value db.
func NewStore(dbs kvdb.FlushableDBProducer, cfg StoreConfig) *Store {
	s := &Store{
		dbs:           dbs,
		cfg:           cfg,
		Instance:      logger.New("gossip-store"),
		prevFlushTime: time.Now(),
		rlp:           rlpstore.Helper{logger.New("rlp")},
	}

	err := table.OpenTables(&s.table, dbs, "gossip")
	if err != nil {
		log.Crit("Failed to open DB", "name", "gossip", "err", err)
	}

	s.initCache()
	s.evm = evmstore.NewStore(dbs, cfg.EVM)
	if cfg.TraceTransactions {
		s.txtrace = txtrace.NewStore(s.table.TransactionTraces)
	}

	if err := s.migrateData(); err != nil {
		s.Log.Crit("Failed to migrate Gossip DB", "err", err)
	}

	return s
}

func (s *Store) initCache() {
	s.cache.Events = s.makeCache(s.cfg.Cache.EventsSize, s.cfg.Cache.EventsNum)
	s.cache.Blocks = s.makeCache(s.cfg.Cache.BlocksSize, s.cfg.Cache.BlocksNum)

	blockHashesNum := s.cfg.Cache.BlocksNum
	blockHashesCacheSize := nominalSize * uint(blockHashesNum)
	s.cache.BlockHashes = s.makeCache(blockHashesCacheSize, blockHashesNum)
	s.cache.BRHashes = s.makeCache(blockHashesCacheSize, blockHashesNum)

	eventsHeadersNum := s.cfg.Cache.EventsNum
	eventsHeadersCacheSize := nominalSize * uint(eventsHeadersNum)
	s.cache.EventsHeaders = s.makeCache(eventsHeadersCacheSize, eventsHeadersNum)

	s.cache.EventIDs = eventid.NewCache(s.cfg.Cache.EventsIDsNum)

	blockEpochStatesNum := s.cfg.Cache.BlockEpochStateNum
	blockEpochStatesSize := nominalSize * uint(blockEpochStatesNum)
	s.cache.BlockEpochStateHistory = s.makeCache(blockEpochStatesSize, blockEpochStatesNum)

	s.cache.LlrBlockVotesIndex = NewVotesCache(s.cfg.Cache.LlrBlockVotesIndexes, s.flushLlrBlockVoteWeight)
	s.cache.LlrEpochVoteIndex = NewVotesCache(s.cfg.Cache.LlrEpochVotesIndexes, s.flushLlrEpochVoteWeight)
}

// Close closes underlying database.
func (s *Store) Close() {
	setnil := func() interface{} {
		return nil
	}

	if s.snapshotedEVMDB != nil {
		s.snapshotedEVMDB.Release()
	}
	_ = table.CloseTables(&s.table)
	table.MigrateTables(&s.table, nil)
	table.MigrateCaches(&s.cache, setnil)

	if s.txtrace != nil {
		s.txtrace.Close()
	}
	_ = s.closeEpochStore()
	s.evm.Close()
}

func (s *Store) IsCommitNeeded() bool {
	// randomize flushing criteria for each epoch so that nodes would desynchronize flushes
	ratio := 900 + randat.RandAt(uint64(s.GetEpoch()))%100
	return s.isCommitNeeded(ratio, ratio)
}

func (s *Store) isCommitNeeded(sc, tc uint64) bool {
	period := s.cfg.MaxNonFlushedPeriod * time.Duration(sc) / 1000
	size := (uint64(s.cfg.MaxNonFlushedSize) / 2) * tc / 1000
	return time.Since(s.prevFlushTime) > period ||
		uint64(s.dbs.NotFlushedSizeEst()) > size
}

// commitEVM commits EVM storage
func (s *Store) commitEVM(flush bool) {
	bs := s.GetBlockState()
	err := s.evm.Commit(bs.LastBlock.Idx, bs.FinalizedStateRoot, flush)
	if err != nil {
		s.Log.Crit("Failed to commit EVM storage", "err", err)
	}
	s.evm.Cap()
}

func (s *Store) cleanCommitEVM() {
	err := s.evm.CleanCommit(s.GetBlockState())
	if err != nil {
		s.Log.Crit("Failed to commit EVM storage", "err", err)
	}
	s.evm.Cap()
}

func (s *Store) GenerateSnapshotAt(root common.Hash, async bool) (err error) {
	err = s.generateSnapshotAt(s.evm, root, true, async)
	if err != nil {
		s.Log.Error("EVM snapshot", "at", root, "err", err)
	} else {
		gen, _ := s.evm.Snaps.Generating()
		s.Log.Info("EVM snapshot", "at", root, "generating", gen)
	}
	return err
}

func (s *Store) generateSnapshotAt(evmStore *evmstore.Store, root common.Hash, rebuild, async bool) (err error) {
	return evmStore.GenerateEvmSnapshot(root, rebuild, async)
}

// Commit changes.
func (s *Store) Commit() error {
	s.FlushBlockEpochState()
	s.FlushHighestLamport()
	s.FlushLastBVs()
	s.FlushLastEV()
	s.FlushLlrState()
	s.cache.LlrBlockVotesIndex.FlushMutated(s.flushLlrBlockVoteWeight)
	s.cache.LlrEpochVoteIndex.FlushMutated(s.flushLlrEpochVoteWeight)
	es := s.getAnyEpochStore()
	if es != nil {
		es.FlushHeads()
		es.FlushLastEvents()
	}
	return s.flushDBs()
}

func (s *Store) flushDBs() error {
	s.prevFlushTime = time.Now()
	flushID := bigendian.Uint64ToBytes(uint64(s.prevFlushTime.UnixNano()))
	return s.dbs.Flush(flushID)
}

func (s *Store) EvmStore() *evmstore.Store {
	return s.evm
}

func (s *Store) TxTraceStore() *txtrace.Store {
	return s.txtrace
}

func (s *Store) CaptureEvmKvdbSnapshot() {
	if s.evm.Snaps == nil {
		return
	}
	gen, err := s.evm.Snaps.Generating()
	if err != nil {
		s.Log.Error("Failed to check EVM snapshot generation", "err", err)
		return
	}
	if gen {
		return
	}
	newEvmKvdbSnap, err := s.evm.EVMDB().GetSnapshot()
	if err != nil {
		s.Log.Error("Failed to initialize frozen KVDB", "err", err)
		return
	}
	if s.snapshotedEVMDB == nil {
		s.snapshotedEVMDB = switchable.Wrap(newEvmKvdbSnap)
	} else {
		old := s.snapshotedEVMDB.SwitchTo(newEvmKvdbSnap)
		// release only after DB is atomically switched
		if old != nil {
			old.Release()
		}
	}
	newStore := s.evm.ResetWithEVMDB(snap2kvdb.Wrap(s.snapshotedEVMDB))
	newStore.Snaps = nil
	root := s.GetBlockState().FinalizedStateRoot
	err = s.generateSnapshotAt(newStore, common.Hash(root), false, false)
	if err != nil {
		s.Log.Error("Failed to initialize EVM snapshot for frozen KVDB", "err", err)
		return
	}
	s.cache.KvdbEvmSnap.Store(newStore)
}

func (s *Store) LastKvdbEvmSnapshot() *evmstore.Store {
	if v := s.cache.KvdbEvmSnap.Load(); v != nil {
		return v.(*evmstore.Store)
	}
	return s.evm
}

/*
 * Utils:
 */

func (s *Store) makeCache(weight uint, size int) *wlru.Cache {
	cache, err := wlru.New(weight, size)
	if err != nil {
		s.Log.Crit("Failed to create LRU cache", "err", err)
		return nil
	}
	return cache
}
