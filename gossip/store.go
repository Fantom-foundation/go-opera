package gossip

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/hashicorp/golang-lru"

	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
	"github.com/Fantom-foundation/go-lachesis/gossip/temporary"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/flushable"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	dbs *flushable.SyncedPool
	cfg StoreConfig

	mainDb kvdb.KeyValueStore
	table  struct {
		// Network tables
		Peers kvdb.KeyValueStore `table:"Z"`

		// Main DAG tables
		Events    kvdb.KeyValueStore `table:"e"`
		Blocks    kvdb.KeyValueStore `table:"b"`
		PackInfos kvdb.KeyValueStore `table:"p"`
		Packs     kvdb.KeyValueStore `table:"P"`
		PacksNum  kvdb.KeyValueStore `table:"n"`

		// general economy tables
		EpochStats kvdb.KeyValueStore `table:"E"`

		// gas power economy tables
		LastEpochHeaders kvdb.KeyValueStore `table:"l"`

		// API-only tables
		BlockHashes     kvdb.KeyValueStore `table:"h"`
		TxPositions     kvdb.KeyValueStore `table:"x"`
		DecisiveEvents  kvdb.KeyValueStore `table:"9"`
		EventLocalTimes kvdb.KeyValueStore `table:"!"`

		TmpDbs kvdb.KeyValueStore `table:"T"`
	}

	EpochDbs *temporary.Dbs

	cache struct {
		Events        *lru.Cache `cache:"-"` // store by pointer
		EventsHeaders *lru.Cache `cache:"-"` // store by pointer
		Blocks        *lru.Cache `cache:"-"` // store by pointer
		PackInfos     *lru.Cache `cache:"-"` // store by value
		EpochStats    *lru.Cache `cache:"-"` // store by value
		TxPositions   *lru.Cache `cache:"-"` // store by pointer
		BlockHashes   *lru.Cache `cache:"-"` // store by pointer
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
		mainDb:   dbs.GetDb("gossip-main"),
		Instance: logger.MakeInstance(),
	}

	table.MigrateTables(&s.table, s.mainDb)

	s.EpochDbs = s.newTmpDbs("epoch", func(ver uint64) (
		db kvdb.KeyValueStore,
		tables interface{},
	) {
		db = s.dbs.GetDb(fmt.Sprintf("gossip-epoch-%d", ver))
		tables = newEpochStore(db)
		return
	})

	s.initCache()

	// for compability with db before commit 591ede6
	s.rmPrefix(s.table.PackInfos, "serverPool")

	return s
}

func (s *Store) newTmpDbs(name string, maker temporary.DbMaker) *temporary.Dbs {
	t := table.New(s.table.TmpDbs, []byte(name))
	dbs := temporary.NewDbs(t, maker)
	dbs.SetName(name)

	return dbs
}

func (s *Store) initCache() {
	s.cache.Events = s.makeCache(s.cfg.EventsCacheSize)
	s.cache.EventsHeaders = s.makeCache(s.cfg.EventsHeadersCacheSize)
	s.cache.Blocks = s.makeCache(s.cfg.BlockCacheSize)
	s.cache.PackInfos = s.makeCache(s.cfg.PackInfosCacheSize)
	s.cache.EpochStats = s.makeCache(s.cfg.EpochStatsCacheSize)
	s.cache.TxPositions = s.makeCache(s.cfg.TxPositionsCacheSize)
	s.cache.BlockHashes = s.makeCache(s.cfg.BlockCacheSize)
}

// Close leaves underlying database.
func (s *Store) Close() {
	setnil := func() interface{} {
		return nil
	}

	table.MigrateTables(&s.table, nil)
	table.MigrateCaches(&s.cache, setnil)

	s.mainDb.Close()
}

// Commit changes.
func (s *Store) Commit(flushID []byte, immediately bool) error {
	if flushID == nil {
		// if flushId not specified, use current time
		buf := bytes.NewBuffer(nil)
		buf.Write([]byte{0xbe, 0xee})                                    // 0xbeee eyecatcher that flushed time
		buf.Write(bigendian.Int64ToBytes(uint64(time.Now().UnixNano()))) // current UnixNano time
		flushID = buf.Bytes()
	}

	if !immediately && !s.dbs.IsFlushNeeded() {
		return nil
	}

	// Flush the DBs
	return s.dbs.Flush(flushID)
}

/*
 * Utils:
 */

// set RLP value
func (s *Store) set(table kvdb.KeyValueStore, key []byte, val interface{}) {
	buf, err := rlp.EncodeToBytes(val)
	if err != nil {
		s.Log.Crit("Failed to encode rlp", "err", err)
	}

	if err := table.Put(key, buf); err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

// get RLP value
func (s *Store) get(table kvdb.KeyValueStore, key []byte, to interface{}) interface{} {
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

func (s *Store) has(table kvdb.KeyValueStore, key []byte) bool {
	res, err := table.Has(key)
	if err != nil {
		s.Log.Crit("Failed to get key", "err", err)
	}
	return res
}

func (s *Store) rmPrefix(t kvdb.KeyValueStore, prefix string) {
	it := t.NewIteratorWithPrefix([]byte(prefix))
	defer it.Release()

	s.dropTable(it, t)
}

func (s *Store) dropTable(it ethdb.Iterator, t kvdb.KeyValueStore) {
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
