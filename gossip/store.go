package gossip

import (
	"bytes"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/hashicorp/golang-lru"

	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/flushable"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/nokeyiserr"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	dbs *flushable.SyncedPool
	cfg StoreConfig

	mainDb kvdb.KeyValueStore
	table  struct {
		Peers            kvdb.KeyValueStore `table:"peer"`
		Events           kvdb.KeyValueStore `table:"event"`
		Blocks           kvdb.KeyValueStore `table:"block"`
		PackInfos        kvdb.KeyValueStore `table:"packinfo"`
		Packs            kvdb.KeyValueStore `table:"pack"`
		PacksNum         kvdb.KeyValueStore `table:"packsnum"`
		LastEpochHeaders kvdb.KeyValueStore `table:"lheaders"`
		EpochStats       kvdb.KeyValueStore `table:"estats"`

		// API-only tables
		BlockHashes kvdb.KeyValueStore `table:"blockh"`
		Receipts    kvdb.KeyValueStore `table:"receipts"`
		TxPositions kvdb.KeyValueStore `table:"txp"`

		// SFC-related tables
		Validators kvdb.KeyValueStore `table:"va"`
		Stakers    kvdb.KeyValueStore `table:"vs"`
		Delegators kvdb.KeyValueStore `table:"de"`

		TmpDbs kvdb.KeyValueStore `table:"tmpdbs"`

		Evm      ethdb.Database
		EvmState state.Database
	}

	cache struct {
		Events           *lru.Cache `cache:"-"` // store by pointer
		EventsHeaders    *lru.Cache `cache:"-"` // store by pointer
		Blocks           *lru.Cache `cache:"-"` // store by pointer
		PackInfos        *lru.Cache `cache:"-"` // store by value
		Receipts         *lru.Cache `cache:"-"` // store by value
		TxPositions      *lru.Cache `cache:"-"` // store by pointer
		EpochStats       *lru.Cache `cache:"-"` // store by value
		LastEpochHeaders *lru.Cache `cache:"-"` // store by pointer
		Stakers          *lru.Cache `cache:"-"` // store by pointer
		Delegators       *lru.Cache `cache:"-"` // store by pointer
	}

	mutexes struct {
		LastEpochHeaders *sync.RWMutex
	}

	tmpDbs

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

	evmTable := nokeyiserr.Wrap(table.New(s.mainDb, []byte("evm_"))) // ETH expects that "not found" is an error
	s.table.Evm = rawdb.NewDatabase(evmTable)
	s.table.EvmState = state.NewDatabaseWithCache(s.table.Evm, 16)

	s.initTmpDbs()
	s.initCache()
	s.initMutexes()

	return s
}

func (s *Store) initCache() {
	s.cache.Events = s.makeCache(s.cfg.EventsCacheSize)
	s.cache.EventsHeaders = s.makeCache(s.cfg.EventsHeadersCacheSize)
	s.cache.Blocks = s.makeCache(s.cfg.BlockCacheSize)
	s.cache.PackInfos = s.makeCache(s.cfg.PackInfosCacheSize)
	s.cache.Receipts = s.makeCache(s.cfg.ReceiptsCacheSize)
	s.cache.TxPositions = s.makeCache(s.cfg.TxPositionsCacheSize)
	s.cache.EpochStats = s.makeCache(s.cfg.EpochStatsCacheSize)
	s.cache.LastEpochHeaders = s.makeCache(s.cfg.LastEpochHeadersCacheSize)
	s.cache.Stakers = s.makeCache(s.cfg.StakersCacheSize)
	s.cache.Delegators = s.makeCache(s.cfg.DelegatorsCacheSize)
}

func (s *Store) initMutexes() {
	s.mutexes.LastEpochHeaders = new(sync.RWMutex)
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

	if immediately || s.dbs.IsFlushNeeded() {
		// Flush trie on the DB
		err := s.table.EvmState.TrieDB().Cap(0)
		if err != nil {
			s.Log.Error("Failed to flush trie DB into main DB", "err", err)
		}
		// Flush the DBs
		return s.dbs.Flush(flushID)
	}

	return nil
}

// StateDB returns state database.
func (s *Store) StateDB(from common.Hash) *state.StateDB {
	db, err := state.New(common.Hash(from), s.table.EvmState)
	if err != nil {
		s.Log.Crit("Failed to open state", "err", err)
	}
	return db
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
		s.Log.Crit("Failed to decode rlp", "err", err)
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
