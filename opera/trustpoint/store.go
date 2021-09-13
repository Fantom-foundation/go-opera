package trustpoint

import (
	"crypto/sha256"

	"github.com/Fantom-foundation/lachesis-base/common/bigendian"
	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"

	"github.com/Fantom-foundation/go-opera/logger"
	"github.com/Fantom-foundation/go-opera/utils/rlpstore"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	GenesisHash hash.Hash

	db    kvdb.Store
	table struct {
		BlockEpochState kvdb.Store `table:"s"`
		Events          kvdb.Store `table:"e"`
		Blocks          kvdb.Store `table:"b"`
		Txs             kvdb.Store `table:"t"`
		Receipts        kvdb.Store `table:"r"`
		RawEvmItems     kvdb.Store `table:"v"`
	}

	rlp rlpstore.Helper
	logger.Instance
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	return NewStore(memorydb.New())
}

// NewStore creates store over key-value db.
func NewStore(db kvdb.Store) *Store {
	s := &Store{
		db:       db,
		Instance: logger.New(),
		rlp:      rlpstore.Helper{logger.New()},
	}

	table.MigrateTables(&s.table, s.db)

	return s
}

// Close leaves underlying database.
func (s *Store) Close() {
	table.MigrateTables(&s.table, nil)

	_ = s.db.Close()
}

func (s *Store) Hash() hash.Hash {
	hasher := sha256.New()
	it := s.db.NewIterator(nil, nil)
	defer it.Release()
	for it.Next() {
		k := it.Key()
		v := it.Value()
		hasher.Write(bigendian.Uint32ToBytes(uint32(len(k))))
		hasher.Write(k)
		hasher.Write(bigendian.Uint32ToBytes(uint32(len(v))))
		hasher.Write(v)
	}
	return hash.BytesToHash(hasher.Sum(nil))
}
