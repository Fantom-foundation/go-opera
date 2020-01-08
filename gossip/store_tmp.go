package gossip

import (
	"fmt"
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
		seq   ringbuf
		maker tmpDbMaker

		sync.Mutex
		logger.Instance
	}

	ringbuf struct {
		Min    uint64
		offset int
		count  int
		seq    [5]*tmpDb // len is a max count
	}

	tmpDbMaker func(ver uint64) (db kvdb.KeyValueStore, tables interface{})
)

func (r *ringbuf) Get(num uint64) *tmpDb {
	i := int(num - r.Min)
	if num < r.Min || i >= r.count {
		return nil
	}

	i = (i + r.offset) % len(r.seq)
	return r.seq[i]
}

func (r *ringbuf) Set(num uint64, val *tmpDb) {
	if r.count >= len(r.seq) {
		panic("no space")
	}

	if r.count == 0 {
		r.Min = num
	}

	i := int(num - r.Min)
	if i != r.count {
		panic(fmt.Sprintf("sequence is broken (set %d to %+v)", num, r))
	}

	i = (i + r.offset) % len(r.seq)
	r.seq[i] = val
	r.count++
}

func (r *ringbuf) Del(num uint64) {
	if num != r.Min {
		panic(fmt.Sprintf("sequence is broken (del %d from %+v)", num, r))
	}

	r.Min++
	r.offset = (r.offset + 1) % len(r.seq)
	r.count--
}

func (s *Store) newTmpDbs(name string, maker tmpDbMaker) *tmpDbs {
	dbs := &tmpDbs{
		store:    table.New(s.table.TmpDbs, []byte(name)),
		maker:    maker,
		Instance: logger.MakeInstance(),
	}
	dbs.SetName(name)
	dbs.loadMin()

	return dbs
}

func (t *tmpDbs) loadMin() {
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

func (t *tmpDbs) saveMin() {
	key := []byte("m")

	err := t.store.Put(key, bigendian.Int64ToBytes(t.seq.Min))
	if err != nil {
		t.Log.Crit("Failed to put key-value", "err", err)
	}
}

func (t *tmpDbs) Get(ver uint64) interface{} {
	t.Lock()
	defer t.Unlock()

	if ver < t.seq.Min {
		return nil
	}

	if tmp := t.seq.Get(ver); tmp != nil {
		return tmp.Tables
	}

	db, tables := t.maker(ver)

	t.seq.Set(ver, &tmpDb{
		Db:     db,
		Tables: tables,
	})

	return tables
}

func (t *tmpDbs) Del(ver uint64) {
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
