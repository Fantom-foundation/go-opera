package posposet

import (
	"fmt"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/Fantom-foundation/go-lachesis/src/posposet/internal"
	"github.com/Fantom-foundation/go-lachesis/src/state"

	"github.com/ethereum/go-ethereum/rlp"
)

// Store is a poset persistent storage working over physical key-value database.
type Store struct {
	historyDB kvdb.Database
	tempDb    kvdb.Database

	table struct {
		Checkpoint     kvdb.Database `table:"checkpoint_"`
		Blocks         kvdb.Database `table:"block_"`
		Event2Block    kvdb.Database `table:"event2block_"`
		SuperFrames    kvdb.Database `table:"sframe_"`
		ConfirmedEvent kvdb.Database `table:"confirmed_"`
		FrameInfos     kvdb.Database `table:"frameinfo_"`
		Balances       state.Database
	}

	epochTable struct {
		Roots       kvdb.Database `table:"roots_"`
		VectorIndex kvdb.Database `table:"vectors_"`
	}

	newTempDb func() kvdb.Database

	logger.Instance
}

// NewStore creates store over key-value db.
func NewStore(db kvdb.Database, newTempDb func() kvdb.Database) *Store {
	s := &Store{
		historyDB: db,
		tempDb:    newTempDb(),
		newTempDb: newTempDb,
		Instance:  logger.MakeInstance(),
	}

	kvdb.MigrateTables(&s.table, s.historyDB)
	kvdb.MigrateTables(&s.epochTable, s.tempDb)
	s.table.Balances = state.NewDatabase(
		s.historyDB.NewTable([]byte("balance_")))

	return s
}

// NewMemStore creates store over memory map.
func NewMemStore() *Store {
	return NewStore(kvdb.NewMemDatabase(), func() kvdb.Database {
		return kvdb.NewMemDatabase()
	})
}

// Close leaves underlying database.
func (s *Store) Close() {
	kvdb.MigrateTables(&s.table, nil)
	kvdb.MigrateTables(&s.epochTable, nil)
	s.historyDB.Close()
	s.tempDb.Close()
}

func (s *Store) pruneTempDb() {
	s.tempDb.Close()
	s.tempDb = s.newTempDb()
	kvdb.MigrateTables(&s.epochTable, s.tempDb)
}

// calcFirstGenesisHash calcs hash of genesis balances.
func calcFirstGenesisHash(balances map[hash.Peer]inter.Stake, time inter.Timestamp) hash.Hash {
	s := NewMemStore()
	defer s.Close()

	if err := s.ApplyGenesis(balances, time); err != nil {
		logger.Get().Fatal(err)
	}
	return s.GetSuperFrame(firstEpoch).PrevEpoch.Hash()
}

// ApplyGenesis stores initial state.
func (s *Store) ApplyGenesis(balances map[hash.Peer]inter.Stake, time inter.Timestamp) error {
	if balances == nil {
		return fmt.Errorf("balances shouldn't be nil")
	}

	sf1 := s.GetSuperFrame(firstEpoch)
	if sf1 != nil {
		if sf1.PrevEpoch.Hash() == calcFirstGenesisHash(balances, time) {
			return nil
		}
		return fmt.Errorf("other genesis has applied already")
	}

	sf := &superFrame{}

	cp := &checkpoint{
		SuperFrameN: firstEpoch,
		TotalCap:    0,
	}

	sf.Members = make(internal.Members, len(balances))

	genesis := s.StateDB(hash.Hash{})
	for addr, balance := range balances {
		if balance == 0 {
			return fmt.Errorf("balance shouldn't be zero")
		}

		genesis.SetBalance(hash.Peer(addr), balance)
		cp.TotalCap += balance

		sf.Members.Add(addr, balance)
	}
	sf.Members = sf.Members.Top()
	cp.NextMembers = sf.Members.Top()

	var err error
	cp.Balances, err = genesis.Commit(true)
	if err != nil {
		return err
	}

	// genesis object
	sf.PrevEpoch.Epoch = cp.SuperFrameN - 1
	sf.PrevEpoch.StateHash = cp.Balances
	sf.PrevEpoch.Time = time
	cp.LastConsensusTime = sf.PrevEpoch.Time

	s.SetSuperFrame(cp.SuperFrameN, sf)
	s.SetCheckpoint(cp)

	return nil
}

/*
 * Utils:
 */

func (s *Store) set(table kvdb.Database, key []byte, val interface{}) {
	buf, err := rlp.EncodeToBytes(val)
	if err != nil {
		s.Fatal(err)
	}

	if err := table.Put(key, buf); err != nil {
		s.Fatal(err)
	}
}

func (s *Store) get(table kvdb.Database, key []byte, to interface{}) interface{} {
	buf, err := table.Get(key)
	if err != nil {
		s.Fatal(err)
	}
	if buf == nil {
		return nil
	}

	err = rlp.DecodeBytes(buf, to)
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
