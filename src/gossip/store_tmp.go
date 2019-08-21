package gossip

import (
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/rlp"

	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
)

type tmpDbs struct {
	min map[string]uint64
	dbs map[string]map[uint64]kvdb.Database

	sync.Mutex
}

func (s *Store) initTmpDbs() {
	s.tmpDbs.min = make(map[string]uint64)
	s.tmpDbs.dbs = make(map[string]map[uint64]kvdb.Database)
	// load mins
	prefix := []byte{}
	err := s.table.TmpDbs.ForEach(prefix, func(key, buf []byte) bool {
		var min uint64
		err := rlp.DecodeBytes(buf, &min)
		if err != nil {
			s.Fatal(err)
		}
		s.tmpDbs.min[string(key)] = min
		return true
	})
	if err != nil {
		s.Fatal(err)
	}
}

func (s *Store) getTmpDb(name string, ver uint64) kvdb.Database {
	s.tmpDbs.Lock()
	defer s.tmpDbs.Unlock()

	if min, ok := s.tmpDbs.min[name]; !ok {
		s.tmpDbs.min[name] = ver
		s.tmpDbs.dbs[name] = make(map[uint64]kvdb.Database)
		s.set(s.table.TmpDbs, []byte(name), &ver)
	} else if ver < min {
		return nil
	}

	if db, ok := s.tmpDbs.dbs[name][ver]; ok {
		return db
	}

	db := s.makeDb(tmpDbName(name, ver))
	s.tmpDbs.dbs[name][ver] = db
	return db
}

func (s *Store) delTmpDb(name string, ver uint64) {
	s.tmpDbs.Lock()
	defer s.tmpDbs.Unlock()

	min, ok := s.tmpDbs.min[name]
	if !ok {
		return
	}

	for i := min; i <= ver; i++ {
		db := s.tmpDbs.dbs[name][i]
		if db != nil {
			db.Close()
			db.Drop()
		}
		delete(s.tmpDbs.dbs[name], i)
	}

	ver += 1
	s.set(s.table.TmpDbs, []byte(name), &ver)
}

func tmpDbName(scope string, ver uint64) string {
	return fmt.Sprintf("tmp_%s_%d", scope, ver)
}
