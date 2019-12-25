package poset

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	lru "github.com/hashicorp/golang-lru"

	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/flushable"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

// Store is a poset persistent storage working over parent key-value database.
type Store struct {
	dbs *flushable.SyncedPool
	cfg StoreConfig

	mainDb kvdb.KeyValueStore
	table  struct {
		Checkpoint     kvdb.KeyValueStore `table:"c"`
		Epochs         kvdb.KeyValueStore `table:"e"`
		ConfirmedEvent kvdb.KeyValueStore `table:"C"`
		FrameInfos     kvdb.KeyValueStore `table:"f"`
	}

	cache struct {
		GenesisHash *common.Hash
		FrameRoots  *lru.Cache `cache:"-"` // store by pointer
	}

	epochDb    kvdb.KeyValueStore
	epochTable struct {
		Roots       kvdb.KeyValueStore `table:"r"`
		VectorIndex kvdb.KeyValueStore `table:"v"`
	}

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(dbs *flushable.SyncedPool, cfg StoreConfig) *Store {
	s := &Store{
		dbs:      dbs,
		cfg:      cfg,
		mainDb:   dbs.GetDb("poset-main"),
		Instance: logger.MakeInstance(),
	}

	table.MigrateTables(&s.table, s.mainDb)

	s.initCache()

	return s
}

func (s *Store) initCache() {
	s.cache.FrameRoots = s.makeCache(s.cfg.Roots)
}

// NewMemStore creates store over memory map.
// Store is always blank.
func NewMemStore() *Store {
	mems := memorydb.NewProducer("")
	dbs := flushable.NewSyncedPool(mems)
	cfg := LiteStoreConfig()

	return NewStore(dbs, cfg)
}

// Close leaves underlying database.
func (s *Store) Close() {
	setnil := func() interface{} {
		return nil
	}

	table.MigrateTables(&s.table, nil)
	table.MigrateCaches(&s.cache, setnil)
	table.MigrateTables(&s.epochTable, nil)
	err := s.mainDb.Close()
	if err != nil {
		s.Log.Crit("Failed to close persistent db", "err", err)
	}

	if s.epochDb == nil {
		return
	}
	err = s.epochDb.Close()
	if err != nil {
		s.Log.Crit("Failed to close epoch db", "err", err)
	}
}

// RecreateEpochDb makes new epoch DB and drops prev.
func (s *Store) RecreateEpochDb(n idx.Epoch) {
	prevDb := s.epochDb
	if prevDb == nil {
		prevDb = s.dbs.GetDb(name(n - 1))
	}

	err := prevDb.Close()
	if err != nil {
		s.Log.Crit("Failed to close epoch db", "err", err)
	}
	prevDb.Drop()

	// Clear full LRU cache.
	if s.cache.FrameRoots != nil {
		s.cache.FrameRoots.Purge()
	}

	s.epochDb = s.dbs.GetDb(name(n))
	table.MigrateTables(&s.epochTable, s.epochDb)
}

func name(n idx.Epoch) string {
	return fmt.Sprintf("poset-epoch-%d", n)
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
