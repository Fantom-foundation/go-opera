package gossip

import (
	"fmt"
	"sync"

	"github.com/Fantom-foundation/go-lachesis/src/common/bigendian"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
)

type (
	tmpDb struct {
		Db     kvdb.KeyValueStore
		Tables interface{}
	}

	tmpDbs struct {
		min map[string]uint64
		seq map[string]map[uint64]tmpDb

		sync.Mutex
	}
)

func (s *Store) initTmpDbs() {
	s.tmpDbs.min = make(map[string]uint64)
	s.tmpDbs.seq = make(map[string]map[uint64]tmpDb)

	// load mins
	it := s.table.TmpDbs.NewIterator()
	defer it.Release()
	for it.Next() {
		min := bigendian.BytesToInt64(it.Value())
		s.tmpDbs.min[string(it.Key())] = min
	}
	if it.Error() != nil {
		s.Log.Crit("Failed to iterate keys", "err", it.Error())
	}
}

func (s *Store) getTmpDb(name string, ver uint64, makeTables func(kvdb.KeyValueStore) interface{}) interface{} {
	s.tmpDbs.Lock()
	defer s.tmpDbs.Unlock()

	if min, ok := s.tmpDbs.min[name]; !ok {
		s.tmpDbs.min[name] = ver
		s.tmpDbs.seq[name] = make(map[uint64]tmpDb)
		err := s.table.TmpDbs.Put([]byte(name), bigendian.Int64ToBytes(ver))
		if err != nil {
			s.Log.Crit("Failed to put key-value", "err", err)
		}
	} else if ver < min {
		return nil
	}

	if _, ok := s.tmpDbs.seq[name]; !ok {
		s.tmpDbs.seq[name] = make(map[uint64]tmpDb)
	} else if tmp, ok := s.tmpDbs.seq[name][ver]; ok {
		return tmp.Tables
	}

	db := s.makeDb(tmpDbName(name, ver))
	tables := makeTables(db)
	s.tmpDbs.seq[name][ver] = tmpDb{
		Db:     db,
		Tables: tables,
	}

	return tables
}

func (s *Store) delTmpDb(name string, ver uint64) {
	s.tmpDbs.Lock()
	defer s.tmpDbs.Unlock()

	min, ok := s.tmpDbs.min[name]
	if !ok || ver < min {
		return
	}

	for i := min; i <= ver; i++ {
		tmp := s.tmpDbs.seq[name][i]
		if tmp.Db != nil {
			tmp.Db.Close()
			tmp.Db.Drop()
		}
		delete(s.tmpDbs.seq[name], i)
	}

	ver += 1
	err := s.table.TmpDbs.Put([]byte(name), bigendian.Int64ToBytes(ver))
	if err != nil {
		s.Log.Crit("Failed to put key-value", "err", err)
	}
}

func tmpDbName(scope string, ver uint64) string {
	return fmt.Sprintf("gossip-%s-%d", scope, ver)
}
