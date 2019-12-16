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
		Peers            kvdb.KeyValueStore `table:"p"`
		Events           kvdb.KeyValueStore `table:"e"`
		Blocks           kvdb.KeyValueStore `table:"b"`
		PackInfos        kvdb.KeyValueStore `table:"p"`
		Packs            kvdb.KeyValueStore `table:"P"`
		PacksNum         kvdb.KeyValueStore `table:"n"`
		LastEpochHeaders kvdb.KeyValueStore `table:"l"`
		EpochStats       kvdb.KeyValueStore `table:"E"`

		// score tables
		ActiveValidationScore      kvdb.KeyValueStore `table:"V"`
		DirtyValidationScore       kvdb.KeyValueStore `table:"v"`
		ActiveOriginationScore     kvdb.KeyValueStore `table:"O"`
		DirtyOriginationScore      kvdb.KeyValueStore `table:"o"`
		BlockParticipation         kvdb.KeyValueStore `table:"m"`
		ValidationScoreCheckpoint  kvdb.KeyValueStore `table:"c"`
		OriginationScoreCheckpoint kvdb.KeyValueStore `table:"C"`

		// API-only tables
		BlockHashes                kvdb.KeyValueStore `table:"h"`
		Receipts                   kvdb.KeyValueStore `table:"r"`
		TxPositions                kvdb.KeyValueStore `table:"x"`
		DelegatorOldRewards        kvdb.KeyValueStore `table:"6"`
		StakerOldRewards           kvdb.KeyValueStore `table:"7"`
		StakerDelegatorsOldRewards kvdb.KeyValueStore `table:"8"`

		// PoI tables
		StakerPOIScore          kvdb.KeyValueStore `table:"s"`
		AddressPOIScore         kvdb.KeyValueStore `table:"a"`
		AddressGasUsed          kvdb.KeyValueStore `table:"g"`
		StakerDelegatorsGasUsed kvdb.KeyValueStore `table:"d"`
		AddressLastTxTime       kvdb.KeyValueStore `table:"X"`
		TotalPOIGasUsed         kvdb.KeyValueStore `table:"U"`

		// SFC-related tables
		Validators kvdb.KeyValueStore `table:"1"`
		Stakers    kvdb.KeyValueStore `table:"2"`
		Delegators kvdb.KeyValueStore `table:"3"`

		TmpDbs kvdb.KeyValueStore `table:"T"`

		Evm      ethdb.Database
		EvmState state.Database
	}

	cache struct {
		Events                     *lru.Cache `cache:"-"` // store by pointer
		EventsHeaders              *lru.Cache `cache:"-"` // store by pointer
		Blocks                     *lru.Cache `cache:"-"` // store by pointer
		PackInfos                  *lru.Cache `cache:"-"` // store by value
		Receipts                   *lru.Cache `cache:"-"` // store by value
		TxPositions                *lru.Cache `cache:"-"` // store by pointer
		EpochStats                 *lru.Cache `cache:"-"` // store by value
		LastEpochHeaders           *lru.Cache `cache:"-"` // store by pointer
		Stakers                    *lru.Cache `cache:"-"` // store by pointer
		Delegators                 *lru.Cache `cache:"-"` // store by pointer
		BlockParticipation         *lru.Cache `cache:"-"` // store by pointer
		BlockHashes                *lru.Cache `cache:"-"` // store by pointer
		ValidationScoreCheckpoint  *lru.Cache `cache:"-"` // store by pointer
		OriginationScoreCheckpoint *lru.Cache `cache:"-"` // store by pointer
	}

	mutexes struct {
		LastEpochHeaders *sync.RWMutex
		IncMutex         *sync.Mutex
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

	evmTable := nokeyiserr.Wrap(table.New(s.mainDb, []byte("M"))) // ETH expects that "not found" is an error
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
	s.cache.BlockParticipation = s.makeCache(256)
	s.cache.BlockHashes = s.makeCache(s.cfg.BlockCacheSize)
	s.cache.ValidationScoreCheckpoint = s.makeCache(4)
	s.cache.OriginationScoreCheckpoint = s.makeCache(4)
}

func (s *Store) initMutexes() {
	s.mutexes.LastEpochHeaders = new(sync.RWMutex)
	s.mutexes.IncMutex = new(sync.Mutex)
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
