package posposet

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/golang-lru"
	"github.com/pkg/errors"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/state"
)

const cacheSize = 500 // TODO: Move it to config later

// Store is a poset persistent storage working over physical key-value database.
type Store struct {
	physicalDB kvdb.Database

	table struct {
		States      kvdb.Database `table:"state_"`
		Frames      kvdb.Database `table:"frame_"`
		Blocks      kvdb.Database `table:"block_"`
		Event2Frame kvdb.Database `table:"event2frame_"`
		Event2Block kvdb.Database `table:"event2block_"`
		Members     kvdb.Database `table:"member_"`
		Balances    state.Database
	}
	cache struct {
		Frames      *lru.Cache `cache:"-"`
		Event2Frame *lru.Cache `cache:"-"`
		Event2Block *lru.Cache `cache:"-"`
	}

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(db kvdb.Database, cached bool) *Store {
	s := &Store{
		physicalDB: db,
		Instance:   logger.MakeInstance(),
	}

	kvdb.MigrateTables(&s.table, s.physicalDB)
	s.table.Balances = state.NewDatabase(kvdb.NewTable(s.physicalDB, "balance_"))

	if cached {
		kvdb.MigrateCaches(&s.cache, func() interface{} {
			c, err := lru.New(cacheSize)
			if err != nil {
				s.Fatal(err)
			}
			return c
		})
	}

	return s
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	db := kvdb.NewMemDatabase()
	return NewStore(db, false)
}

// Close leaves underlying database.
func (s *Store) Close() {
	kvdb.MigrateCaches(&s.cache, nil)
	kvdb.MigrateTables(&s.table, nil)
	s.physicalDB.Close()
}

type sortingBalances struct {
	Addrs    []hash.Peer
	Balances []uint64
}

func (s *sortingBalances) Sort() error {
	if len(s.Addrs) != len(s.Balances) {
		return errors.New("the count of Addrs isn't equal to the count of Balances")
	}

	sort.Sort(s)
	return nil
}

func (s *sortingBalances) Len() int {
	return len(s.Addrs)
}

func (s *sortingBalances) Less(i, j int) bool {
	if s.Balances[i] != s.Balances[j] {
		return s.Balances[i] > s.Balances[j]
	}

	return bytes.Compare(s.Addrs[i].Bytes(), s.Addrs[j].Bytes()) < 0
}

func (s *sortingBalances) Swap(i, j int) {
	s.Addrs[i], s.Addrs[j] = s.Addrs[j], s.Addrs[i]
	s.Balances[i], s.Balances[j] = s.Balances[j], s.Balances[i]
}

const countPeerWithMaxBalance = 30

// ApplyGenesis stores initial state.
func (s *Store) ApplyGenesis(balances map[hash.Peer]uint64) error {
	if balances == nil {
		return fmt.Errorf("sortingBalances shouldn't be nil")
	}

	st := s.GetState()
	if st != nil {
		if st.Genesis == genesisHash(balances) {
			return nil
		}
		return fmt.Errorf("other genesis has applied already")
	}

	st = &State{
		lastFinishedFrameN: 0,
		TotalCap:           0,
	}

	sortBalances := &sortingBalances{
		Addrs:    make([]hash.Peer, 0, len(balances)),
		Balances: make([]uint64, 0, len(balances)),
	}

	genesis := s.StateDB(hash.Hash{})
	for addr, balance := range balances {
		genesis.SetBalance(hash.Peer(addr), balance)
		st.TotalCap += balance

		sortBalances.Addrs = append(sortBalances.Addrs, addr)
		sortBalances.Balances = append(sortBalances.Balances, balance)
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

	err = sortBalances.Sort()
	if err != nil {
		return err
	}

	s.SetMember(0, sortBalances.Addrs[:countPeerWithMaxBalance-1])
	return nil
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
