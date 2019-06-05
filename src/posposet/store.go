package posposet

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/golang-lru"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

const cacheSize = 500 // TODO: Move it to config later

// Store is a poset persistent storage working over physical key-value database.
// TODO: cache tables with LRU.
type Store struct {
	PhysicalDB kvdb.Database `store:"-"`

	States      kvdb.Database `table:"state_"`
	Frames      kvdb.Database `table:"frame_"`
	Blocks      kvdb.Database `table:"block_"`
	Event2frame kvdb.Database `table:"event2frame_"`

	framesCache      *lru.Cache
	event2frameCache *lru.Cache

	balances state.Database `store:"-"` // trie

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(db kvdb.Database, cached bool) *Store {
	s := &Store{
		PhysicalDB: db,
		Instance:   logger.MakeInstance(),
	}

	s.Open()
	if cached {
		s.initCache()
	}

	return s
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	db := kvdb.NewMemDatabase()
	return NewStore(db, false)
}

func (s *Store) initCache() {
	cache := func() *lru.Cache {
		c, err := lru.New(cacheSize)
		if err != nil {
			s.Fatal(err)
		}
		return c
	}

	s.framesCache = cache()
	s.event2frameCache = cache()
}

// Open populate underlying database.
// Open() receiver type method satisfies an abstract database interface
func (s *Store) Open() {
	kvdb.MigrateTables(s, s.PhysicalDB, true)
	s.balances = state.NewDatabase(kvdb.NewTable(s.PhysicalDB, "balance_"))
}

// Close leaves underlying database.
// Close() receiver type method satisfies the abstract database interface
func (s *Store) Close() {
	kvdb.MigrateTables(s, nil, false)
	s.PhysicalDB.Close()

	s.framesCache.Purge()
	s.event2frameCache.Purge()
}

// ApplyGenesis stores initial state.
func (s *Store) ApplyGenesis(balances map[hash.Peer]uint64) error {
	if balances == nil {
		return fmt.Errorf("balances shouldn't be nil")
	}

	st := s.GetState()
	if st != nil {
		if st.Genesis == genesisHash(balances) {
			return nil
		}
		return fmt.Errorf("other genesis has applied already")
	}

	st = &State{
		LastFinishedFrameN: 0,
		TotalCap:           0,
	}

	genesis := s.StateDB(hash.Hash{})
	for addr, balance := range balances {
		genesis.SetBalance(hash.Peer(addr), balance)
		st.TotalCap += balance
	}

	if st.TotalCap < uint64(len(balances)) {
		return fmt.Errorf("balance shouldn't be zero")
	}

	var err error
	st.Genesis, err = genesis.Commit(true)
	if err != nil {
		return err
	}

	s.SetState(st)
	return nil
}

/*
 * Utils:
 */

func (s *Store) set(table kvdb.Database, key []byte, val proto.Message) {
	var pbf proto.Buffer

	if err := pbf.Marshal(val); err != nil {
		panic(err)
	}

	if err := table.Put(key, pbf.Bytes()); err != nil {
		panic(err)
	}
}

func (s *Store) get(table kvdb.Database, key []byte, to proto.Message) proto.Message {
	buf, err := table.Get(key)
	if err != nil {
		panic(err)
	}
	if buf == nil {
		return nil
	}

	err = proto.Unmarshal(buf, to)
	if err != nil {
		panic(err)
	}
	return to
}

func (s *Store) has(table kvdb.Database, key []byte) bool {
	res, err := table.Has(key)
	if err != nil {
		panic(err)
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
