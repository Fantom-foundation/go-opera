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

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
	lru "github.com/hashicorp/golang-lru"

	"github.com/Fantom-foundation/go-opera/gossip/evmstore"
	"github.com/Fantom-foundation/go-opera/logger"
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
		PackInfos  kvdb.Store `table:"p"`
		Packs      kvdb.Store `table:"P"`
		PacksNum   kvdb.Store `table:"n"`

		// API-only
		BlockHashes kvdb.Store `table:"B"`
	}

	epochStore atomic.Value

	cache struct {
		Events        *lru.Cache   `cache:"-"` // store by pointer
		EventsHeaders *lru.Cache   `cache:"-"` // store by pointer
		Blocks        *lru.Cache   `cache:"-"` // store by pointer
		PackInfos     *lru.Cache   `cache:"-"` // store by value
		BlockHashes   *lru.Cache   `cache:"-"` // store by pointer
		BlockState    atomic.Value // store by pointer
		EpochState    atomic.Value // store by pointer
	}

	mutex struct {
		Inc sync.Mutex
	}

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
	s.cache.Events = s.makeCache(s.cfg.EventsCacheSize)
	s.cache.EventsHeaders = s.makeCache(s.cfg.EventsHeadersCacheSize)
	s.cache.Blocks = s.makeCache(s.cfg.BlockCacheSize)
	s.cache.PackInfos = s.makeCache(s.cfg.PackInfosCacheSize)
	s.cache.BlockHashes = s.makeCache(s.cfg.BlockCacheSize)
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
	err := s.app.Commit()
	if err != nil {
		return err
	}
	return s.dbs.Flush(flushID)
}

/*
 * Utils:
 */

// set RLP value
func (s *Store) set(table kvdb.Store, key []byte, val interface{}) {
	buf, err := rlp.EncodeToBytes(val)
	if err != nil {
		s.Log.Crit("Failed to encode rlp", "err", err)
	}

	if err := table.Put(key, buf); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// get RLP value
func (s *Store) get(table kvdb.Store, key []byte, to interface{}) interface{} {
	buf, err := table.Get(key)
	if err != nil {
		s.Log.Crit("Failed to get key-value", "err", err)
	}
	if buf == nil {
		return nil
	}

	err = rlp.DecodeBytes(buf, to)
	if err != nil {
		s.Log.Crit("Failed to decode rlp", "err", err, "size", len(buf))
	}
	return to
}

func (s *Store) has(table kvdb.Store, key []byte) bool {
	res, err := table.Has(key)
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}
	return res
}

func (s *Store) rmPrefix(t kvdb.Store, prefix string) {
	it := t.NewIteratorWithPrefix([]byte(prefix))
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

func (s *Store) makeCache(size int) *lru.Cache {
	if size <= 0 {
		return nil
	}

	cache, err := lru.New(size)
	if err != nil {
		s.Log.Crit("Error create LRU cache", "err", err)
		return nil
	}
	return cache
}
