package gossip

import (
	"sync"

	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

type (
	// tmpDb is a dinamic Db
	tmpDb struct {
		Db     kvdb.KeyValueStore
		Tables interface{}
	}

	// tmpDbs is a named sequence of tmpDb
	tmpDbs struct {
		store kvdb.KeyValueStore
		min   uint64
		seq   map[uint64]tmpDb
		maker tmpDbMaker

		sync.Mutex
		logger.Instance
	}

	tmpDbMaker func(ver uint64) (db kvdb.KeyValueStore, tables interface{})
)

func (s *Store) newTmpDbs(name string, maker tmpDbMaker) *tmpDbs {
	dbs := &tmpDbs{
		store:    table.New(s.table.TmpDbs, []byte(name)),
		seq:      make(map[uint64]tmpDb),
		maker:    maker,
		Instance: logger.MakeInstance(),
	}
	dbs.SetName(name)
	dbs.loadMin()

	return dbs
}

func (dbs *tmpDbs) loadMin() {
	key := []byte("m")

	buf, err := dbs.store.Get(key)
	if err != nil {
		dbs.Log.Crit("Failed to get key-value", "err", err)
	}
	if buf == nil {
		return
	}

	dbs.min = bigendian.BytesToInt64(buf)
}

func (dbs *tmpDbs) saveMin() {
	key := []byte("m")

	err := dbs.store.Put(key, bigendian.Int64ToBytes(dbs.min))
	if err != nil {
		dbs.Log.Crit("Failed to put key-value", "err", err)
	}
}

func (dbs *tmpDbs) Get(ver uint64) interface{} {
	dbs.Lock()
	defer dbs.Unlock()

	if ver < dbs.min {
		return nil
	}

	if tmp, ok := dbs.seq[ver]; ok {
		return tmp.Tables
	}

	db, tables := dbs.maker(ver)

	dbs.seq[ver] = tmpDb{
		Db:     db,
		Tables: tables,
	}

	return tables
}

func (dbs *tmpDbs) Del(ver uint64) {
	dbs.Lock()
	defer dbs.Unlock()

	if ver < dbs.min {
		return
	}

	for i := dbs.min; i <= ver; i++ {
		tmp := dbs.seq[i]
		if tmp.Db != nil {
			tmp.Db.Close()
			tmp.Db.Drop()
		}
		delete(dbs.seq, i)
	}

	ver++
	dbs.min = ver
	dbs.saveMin()
}
