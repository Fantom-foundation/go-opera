package gossip

import (
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
	"github.com/Fantom-foundation/go-opera/gossip/sfcapi"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/utils"
	"github.com/Fantom-foundation/go-opera/utils/rlpstore"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	dbs kvdb.FlushableDBProducer
	cfg StoreConfig

	async *asyncStore

	mainDB kvdb.Store
	evm    *evmstore.Store
	sfcapi *sfcapi.Store
	table  struct {
		Version kvdb.Store `table:"_"`

		// Main DAG tables
		BlockEpochState        kvdb.Store `table:"D"`
		BlockEpochStateHistory kvdb.Store `table:"h"`
		Events                 kvdb.Store `table:"e"`
		Blocks                 kvdb.Store `table:"b"`
		EpochBlocks            kvdb.Store `table:"P"`
		Genesis                kvdb.Store `table:"g"`

		// P2P-only
		HighestLamport kvdb.Store `table:"l"`

		// Network version
		NetworkVersion kvdb.Store `table:"V"`

		// API-only
		BlockHashes kvdb.Store `table:"B"`
		SfcAPI      kvdb.Store `table:"S"`

		LlrState           kvdb.Store `table:"!"`
		LlrBlockResults    kvdb.Store `table:"@"`
		LlrEpochResults    kvdb.Store `table:"#"`
		LlrBlockVotes      kvdb.Store `table:"$"`
		LlrBlockVotesIndex kvdb.Store `table:"%"`
		LlrEpochVotes      kvdb.Store `table:"^"`
		LlrEpochVoteIndex  kvdb.Store `table:"&"`
		LlrLastBlockVotes  kvdb.Store `table:"*"`
		LlrLastEpochVote   kvdb.Store `table:"("`
	}

	prevFlushTime time.Time

	epochStore atomic.Value

	cache struct {
		Events          *wlru.Cache  `cache:"-"` // store by pointer
		EventsHeaders   *wlru.Cache  `cache:"-"` // store by pointer
		Blocks          *wlru.Cache  `cache:"-"` // store by pointer
		BlockHashes     *wlru.Cache  `cache:"-"` // store by pointer
		EvmBlocks       *wlru.Cache  `cache:"-"` // store by pointer
		BlockEpochState atomic.Value // store by value
		HighestLamport  atomic.Value // store by value
		LastBVs         atomic.Value
		LastEV          atomic.Value
		LlrState        atomic.Value
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
	mainDB, err := dbs.OpenDB("gossip")
	if err != nil {
		log.Crit("Filed to open DB", "name", "gossip", "err", err)
	}
	asyncDB, err := dbs.OpenDB("gossip-async")
	if err != nil {
		log.Crit("Filed to open DB", "name", "gossip-async", "err", err)
	}
	s := &Store{
		dbs:           dbs,
		cfg:           cfg,
		async:         newAsyncStore(asyncDB),
		mainDB:        mainDB,
		Instance:      logger.New("gossip-store"),
		prevFlushTime: time.Now(),
		rlp:           rlpstore.Helper{logger.New("rlp")},
	}

	table.MigrateTables(&s.table, s.mainDB)

	s.initCache()
	s.evm = evmstore.NewStore(s.mainDB, cfg.EVM)
	s.sfcapi = sfcapi.NewStore(s.table.SfcAPI)

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

	eventsHeadersNum := s.cfg.Cache.EventsNum
	eventsHeadersCacheSize := nominalSize * uint(eventsHeadersNum)
	s.cache.EventsHeaders = s.makeCache(eventsHeadersCacheSize, eventsHeadersNum)
}

// Close closes underlying database.
func (s *Store) Close() {
	setnil := func() interface{} {
		return nil
	}

	table.MigrateTables(&s.table, nil)
	table.MigrateCaches(&s.cache, setnil)

	_ = s.mainDB.Close()
	s.async.Close()
	s.sfcapi.Close()
	_ = s.closeEpochStore()
}

func (s *Store) IsCommitNeeded(epochSealing bool) bool {
	period := s.cfg.MaxNonFlushedPeriod
	size := s.cfg.MaxNonFlushedSize / 2
	if epochSealing {
		period /= 2
		size /= 2
	}
	return time.Since(s.prevFlushTime) > period ||
		s.dbs.NotFlushedSizeEst() > size
}

// commitEVM commits EVM storage
func (s *Store) commitEVM(genesis bool) {
	err := s.evm.Commit(s.GetBlockState(), genesis)
	if err != nil {
		s.Log.Crit("Failed to commit EVM storage", "err", err)
	}
	s.evm.Cap(s.cfg.MaxNonFlushedSize/3, s.cfg.MaxNonFlushedSize/4)
}

func (s *Store) EvmSnapshotAt(root common.Hash) (err error) {
	start := time.Now()
	defer func() {
		now := time.Now()
		if err != nil {
			s.Log.Error("EVM snapshot", "at", root, "err", err, "t", utils.PrettyDuration(now.Sub(start)))
		} else {
			s.Log.Info("EVM snapshot", "at", root, "t", utils.PrettyDuration(now.Sub(start)))
		}
	}()

	// DB is being flushed in a middle of this call to limit memory usage of snapshot building
	res := make(chan error)
	go func() {
		res <- s.EvmStore().CreateEvmSnapshot(root)
	}()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if s.IsCommitNeeded(false) {
				_ = s.Commit()
			}
		case err = <-res:
			_ = s.Commit()
			return
		}
	}
}

// Commit changes.
func (s *Store) Commit() error {
	s.prevFlushTime = time.Now()
	flushID := bigendian.Uint64ToBytes(uint64(time.Now().UnixNano()))
	// Flush the DBs
	s.FlushBlockEpochState()
	s.FlushHighestLamport()
	s.FlushLastBVs()
	s.FlushLastEV()
	s.FlushLlrState()
	es := s.getAnyEpochStore()
	if es != nil {
		es.FlushHeads()
		es.FlushLastEvents()
	}
	return s.dbs.Flush(flushID)
}

func (s *Store) EvmStore() *evmstore.Store {
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
