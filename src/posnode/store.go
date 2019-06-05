package posnode

import (
	"github.com/golang/protobuf/proto"

	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

// Store is a node persistent storage working over physical key-value database.
type Store struct {
	PhysicalDB kvdb.Database `store:"-"`

	Peers       kvdb.Database `table:"peer_"`
	PeersTop    kvdb.Database `table:"top_peers_"`
	PeerHeights kvdb.Database `table:"peer_height_"`

	Events kvdb.Database `table:"event_"`
	Hashes kvdb.Database `table:"hash_"`

	Nonce       kvdb.Database `table:"nonce_"`
	NonceTxs    kvdb.Database `table:"nonce_tx_"`
	NonceEvents kvdb.Database `table:"nonce_event_"`

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(db kvdb.Database) *Store {
	s := &Store{
		PhysicalDB: db,
		Instance:   logger.MakeInstance(),
	}
	s.Open()
	return s
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	db := kvdb.NewMemDatabase()
	return NewStore(db)
}

// Open populate underlying database.
// Open() receiver type method satisfies an abstract database interface
func (s *Store) Open() {
	kvdb.MigrateTables(s, s.PhysicalDB, true)
}

// Close leaves underlying database.
// Close() receiver type method satisfies the abstract database interface
func (s *Store) Close() {
	kvdb.MigrateTables(s, s.PhysicalDB, false)
	s.PhysicalDB.Close()
}

/*
 * Utils:
 */

func (s *Store) set(table kvdb.Database, key []byte, val proto.Message) {
	var pbf proto.Buffer

	if err := pbf.Marshal(val); err != nil {
		s.Fatal(err)
	}

	if err := table.Put(key, pbf.Bytes()); err != nil {
		s.Fatal(err)
	}
}

func (s *Store) get(table kvdb.Database, key []byte, to proto.Message) proto.Message {
	buf, err := table.Get(key)
	if err != nil {
		s.Fatal(err)
	}
	if buf == nil {
		return nil
	}

	err = proto.Unmarshal(buf, to)
	if err != nil {
		s.Fatal(err)
	}
	return to
}

func (s *Store) has(table kvdb.Database, key []byte) bool {
	res, err := table.Has(key)
	if err != nil {
		s.Fatal(err)
	}
	return res
}

func intToBytes(n uint64) []byte {
	var res [8]byte
	for i := 0; i < len(res); i++ {
		res[i] = byte(n)
		n = n >> 8
	}
	return res[:]
}

func bytesToInt(b []byte) uint64 {
	var res uint64
	for i := 0; i < len(b); i++ {
		res += uint64(b[i]) << uint(i*8)
	}
	return res
}
