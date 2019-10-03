package main

import (
	"bytes"
	"errors"
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/flushable"
	"time"
)

type RegisteredDbs struct {
	wrappers map[string]kvdb.FlushableKeyValueStore
	bareDbs map[string]kvdb.KeyValueStore

	queuedDrops map[string]struct{}

	prevFlushTime time.Time
}

func (dbs *RegisteredDbs) Drop(name string) {
	dbs.queuedDrops[name] = struct{}{}
}

func (dbs *RegisteredDbs) Register(name string, db kvdb.KeyValueStore) {
	wrapper := flushable.New(db)
	wrapper.SetDropper(func() {
		dbs.Drop(name)
	})

	dbs.bareDbs[name] = db
	dbs.wrappers[name] = wrapper
	delete(dbs.queuedDrops, name)
}

func (dbs *RegisteredDbs) Flush(id hash.Event) error {
	key := []byte("mark")

	// dirty flag
	for _, db := range dbs.bareDbs {
		marker := bytes.NewBuffer(nil)
		prev, err := db.Get(key)
		if err != nil {
			return err
		}
		if prev == nil {
			return errors.New("not found prev flushed state marker")
		}

		marker.Write([]byte("dirty"))
		marker.Write(prev)
		marker.Write([]byte("->"))
		marker.Write(id.Bytes())
		err = db.Put(key, marker.Bytes())
		if err != nil {
			return err
		}
	}
	// flush (along with clean flag)
	for _, db := range dbs.wrappers {
		db.Flush() // flush data
		err := db.Put(key, id.Bytes())
		if err != nil {
			return err
		}
		db.Flush() // flush clean flag
	}

	// drop old DBs
	for name := range dbs.queuedDrops {
		db := dbs.bareDbs[name]
		if db == nil {
			// ???? we should register all the DBs at startup
			continue
		}
		db.Drop()
	}

	dbs.prevFlushTime = time.Now()
	return nil
}

func (dbs *RegisteredDbs) FlushIfNeeded(id hash.Event) bool {
	if time.Since(dbs.prevFlushTime) > 10 * time.Minute {
		dbs.Flush(id)
		return true
	}

	totalNotFlushed := 0
	for _, db := range dbs.wrappers {
		totalNotFlushed += db.NotFlushedSizeEst()
	}

	if totalNotFlushed > 100 * 1024 * 1024 {
		dbs.Flush(id)
		return true
	}
	return false
}

// call on startup, after all dbs are registered
func (dbs *RegisteredDbs) CheckDbsSynced() error {
	key := []byte("mark")
	var prevId *hash.Event
	for _, db := range dbs.bareDbs {
		mark, err := db.Get(key)
		if err != nil {
			return err
		}
		if bytes.HasPrefix(mark, []byte("dirty")) {
			return errors.New("dirty")
		}
		eventId := hash.BytesToEvent(mark)
		if prevId == nil {
			prevId  = &eventId
		}
		if eventId != *prevId {
			return errors.New("not synced")
		}
	}
	return nil
}
