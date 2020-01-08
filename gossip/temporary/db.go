package temporary

import (
	"sync"

	"github.com/Fantom-foundation/go-lachesis/common/bigendian"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

type (
	// Dbs is a named sequence of temporary databases and related object (tables).
	Dbs struct {
		store kvdb.KeyValueStore
		seq   ringbuf
		maker DbMaker

		sync.Mutex
		logger.Instance
	}

	// DbMaker makes temporary database and related object (tables).
	DbMaker func(ver uint64) (db kvdb.KeyValueStore, tables interface{})

	// db is a pair of temporary database and related object (tables).
	pair struct {
		Db     kvdb.KeyValueStore
		Tables interface{}
	}
)

// NewDbs constructor.
func NewDbs(table kvdb.KeyValueStore, maker DbMaker) *Dbs {
	dbs := &Dbs{
		store:    table,
		maker:    maker,
		Instance: logger.MakeInstance(),
	}
	dbs.loadMin()

	return dbs

}

// Get returns related object (tables) of temporary db.
func (t *Dbs) Get(ver uint64) interface{} {
	t.Lock()
	defer t.Unlock()

	if ver < t.seq.Min {
		return nil
	}

	if tmp := t.seq.Get(ver); tmp != nil {
		return tmp.Tables
	}

	p := new(pair)
	p.Db, p.Tables = t.maker(ver)
	t.seq.Set(ver, p)

	return p.Tables
}

// Del removes temporary db.
func (t *Dbs) Del(ver uint64) {
	t.Lock()
	defer t.Unlock()

	if ver < t.seq.Min {
		return
	}

	for i := t.seq.Min; i <= ver; i++ {
		tmp := t.seq.Get(i)
		if tmp != nil {
			tmp.Db.Close()
			tmp.Db.Drop()
		}
		t.seq.Del(i)
	}

	t.saveMin()
}

func (t *Dbs) loadMin() {
	key := []byte("m")

	buf, err := t.store.Get(key)
	if err != nil {
		t.Log.Crit("Failed to get key-value", "err", err)
	}
	if buf == nil {
		return
	}

	t.seq.Min = bigendian.BytesToInt64(buf)
}

func (t *Dbs) saveMin() {
	key := []byte("m")

	err := t.store.Put(key, bigendian.Int64ToBytes(t.seq.Min))
	if err != nil {
		t.Log.Crit("Failed to put key-value", "err", err)
	}
}
