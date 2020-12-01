package gossip

import (
	"bytes"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/flushable"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/Fantom-foundation/lachesis-base/utils/wlru"
	"github.com/ethereum/go-ethereum/ethdb"

	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/utils/rlpstore"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	dbs *flushable.SyncedPool
	cfg StoreConfig

	async *asyncStore

	mainDB kvdb.Store
	evm    *evmstore.Store
	table  struct {
		Version kvdb.Store `table:"_"`

		// Main DAG tables
		BlockState kvdb.Store `table:"z"`
		EpochState kvdb.Store `table:"D"`
		Events     kvdb.Store `table:"e"`
		Blocks     kvdb.Store `table:"b"`
		Packs      kvdb.Store `table:"P"`
		PacksNum   kvdb.Store `table:"n"`
		Genesis    kvdb.Store `table:"g"`

		// API-only
		BlockHashes kvdb.Store `table:"B"`
	}

	epochStore atomic.Value

	cache struct {
		Events        *wlru.Cache  `cache:"-"` // store by pointer
		EventsHeaders *wlru.Cache  `cache:"-"` // store by pointer
		Blocks        *wlru.Cache  `cache:"-"` // store by pointer
		BlockHashes   *wlru.Cache  `cache:"-"` // store by pointer
		BlockState    atomic.Value // store by pointer
		EpochState    atomic.Value // store by pointer
	}

	mutex struct {
		Inc sync.Mutex
	}

	rlp rlpstore.Helper

	logger.Instance
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	mems := memorydb.NewProducer("")
	dbs := flushable.NewSyncedPool(mems)
	cfg := LiteStoreConfig()

	return NewStore(dbs, cfg)
}

// NewStore creates store over key-value db.
func NewStore(dbs *flushable.SyncedPool, cfg StoreConfig) *Store {
	s := &Store{
		dbs:      dbs,
		cfg:      cfg,
		async:    newAsyncStore(dbs),
		mainDB:   dbs.GetDb("gossip"),
		Instance: logger.MakeInstance(),
	}

	table.MigrateTables(&s.table, s.mainDB)

	s.initCache()
	s.evm = evmstore.NewStore(s.mainDB, cfg.EVM)

	return s
}

func (s *Store) initCache() {
	s.cache.Events = s.makeCache(s.cfg.Cache.EventsSize, s.cfg.Cache.EventsNum)
	s.cache.Blocks = s.makeCache(s.cfg.Cache.BlocksSize, s.cfg.Cache.BlocksNum)
	s.cache.EventsHeaders = s.makeCache(uint(s.cfg.Cache.EventsHeadersNum), s.cfg.Cache.EventsHeadersNum)
	s.cache.BlockHashes = s.makeCache(uint(s.cfg.Cache.BlocksNum), s.cfg.Cache.BlocksNum)
}

// Close leaves underlying database.
func (s *Store) Close() {
	setnil := func() interface{} {
		return nil
	}

	table.MigrateTables(&s.table, nil)
	table.MigrateCaches(&s.cache, setnil)

	s.mainDB.Close()
	s.async.Close()
}

// Commit changes.
func (s *Store) Commit(flushID []byte, immediately bool) error {
	if flushID == nil {
		// if flushId not specified, use current time
		buf := bytes.NewBuffer(nil)
		buf.Write([]byte{0xbe, 0xee})                                     // 0xbeee eyecatcher that flushed time
		buf.Write(bigendian.Uint64ToBytes(uint64(time.Now().UnixNano()))) // current UnixNano time
		flushID = buf.Bytes()
	}

	if !immediately && !s.dbs.IsFlushNeeded() {
		return nil
	}

	// Flush the DBs
	s.FlushBlockState()
	s.FlushEpochState()
	err := s.evm.Commit()
	if err != nil {
		return err
	}
	return s.dbs.Flush(flushID)
}

/*
 * Utils:
 */

func (s *Store) rmPrefix(t kvdb.Store, prefix string) {
	it := t.NewIterator([]byte(prefix), nil)
	defer it.Release()

	s.dropTable(it, t)
}

func (s *Store) dropTable(it ethdb.Iterator, t kvdb.Store) {
	keys := make([][]byte, 0, 500) // don't write during iteration

	for it.Next() {
		keys = append(keys, it.Key())
	}

	for i := range keys {
		err := t.Delete(keys[i])
		if err != nil {
			s.Log.Crit("Failed to erase key-value", "err", err)
		}
	}
}

func (s *Store) makeCache(weight uint, size int) *wlru.Cache {
	cache, err := wlru.New(weight, size)
	if err != nil {
		s.Log.Crit("Error create LRU cache", "err", err)
		return nil
	}
	return cache
}
