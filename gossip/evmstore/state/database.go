package state

import (
	"errors"
	"fmt"

	//"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ledgerwatch/erigon-lib/kv"
	erigonethdb "github.com/ledgerwatch/erigon/ethdb"

	lru "github.com/hashicorp/golang-lru"

	"github.com/VictoriaMetrics/fastcache"
)

const (
	// Number of codehash->size associations to keep.
	codeSizeCacheSize = 100000

	// Cache size granted for caching clean code.
	codeCacheSize = 64 * 1024 * 1024
)

// Database wraps access to tries and contract code.
type Database interface {

	// ContractCode retrieves a particular contract's code.
	ContractCode(addrHash, codeHash common.Hash) ([]byte, error)

	// ContractCodeSize retrieves a particular contracts code's size.
	ContractCodeSize(addrHash, codeHash common.Hash) (int, error)
}

// NewDatabase creates a backing store for state. The returned database is safe for
// concurrent use, but does not retain any recent trie nodes in memory. To keep some
// historical state in memory, use the NewDatabaseWithConfig constructor.
// TODO add genesisKV to handle ContractCode and other methods
func NewDatabase(kv erigonethdb.Database) Database {
	csc, _ := lru.New(codeSizeCacheSize)
	return &cachingDB{
		kv:            kv,
		codeSizeCache: csc,
		codeCache:     fastcache.New(codeCacheSize),
	}
}

type cachingDB struct {
	kv            erigonethdb.Database
	codeSizeCache *lru.Cache
	codeCache     *fastcache.Cache
}

// ContractCode retrieves a particular contract's code.
func (db *cachingDB) ContractCode(addrHash, codeHash common.Hash) ([]byte, error) {
	if code := db.codeCache.Get(nil, codeHash.Bytes()); len(code) > 0 {
		return code, nil
	}
	code, err := db.kv.GetOne(kv.Code, codeHash.Bytes())
	if err != nil {
		return nil, fmt.Errorf("can not fetch contract code: err: %w", err)
	}
	if len(code) > 0 {
		db.codeCache.Set(codeHash.Bytes(), code)
		db.codeSizeCache.Add(codeHash, len(code))
		return code, nil
	}
	return nil, errors.New("not found")
}

// ContractCodeSize retrieves a particular contracts code's size.
func (db *cachingDB) ContractCodeSize(addrHash, codeHash common.Hash) (int, error) {
	if cached, ok := db.codeSizeCache.Get(codeHash); ok {
		return cached.(int), nil
	}
	code, err := db.ContractCode(addrHash, codeHash)
	return len(code), err
}
