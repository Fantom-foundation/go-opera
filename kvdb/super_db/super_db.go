package super_db

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	
	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/flushable"
	"github.com/Fantom-foundation/go-lachesis/kvdb/leveldb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
)

type SuperDb struct {
	pathes   map[string]string
	wrappers map[string]*flushable.Flushable
	bareDbs  map[string]kvdb.KeyValueStore

	queuedDrops map[string]struct{}

	prevFlushTime time.Time

	datadir string
}

func New(datadir string) *SuperDb {
	dirs, err := ioutil.ReadDir("./")
	if err != nil {
		return nil
	}

	sdb := &SuperDb{
		pathes:   make(map[string]string),
		wrappers: make(map[string]*flushable.Flushable),
		bareDbs:  make(map[string]kvdb.KeyValueStore),

		queuedDrops: make(map[string]struct{}),
		datadir:     datadir,
	}

	for _, f := range dirs {
		path := f.Name()
		if f.IsDir() && strings.HasSuffix(path, "-ldb") {
			name := strings.TrimSuffix(path, "-ldb")
			sdb.registerExisting(name, path)
		}
	}
	return sdb
}

func (sdb *SuperDb) registerExisting(name, path string) kvdb.KeyValueStore {
	db, err := openDb(path)
	if err != nil {
		// TODO log
		return nil
	}
	wrapper := flushable.New(db)

	sdb.pathes[name] = path
	sdb.bareDbs[name] = db
	sdb.wrappers[name] = wrapper
	delete(sdb.queuedDrops, name)
	return wrapper
}

func (sdb *SuperDb) registerNew(name, path string) kvdb.KeyValueStore {
	wrapper := flushable.New(memorydb.New())

	sdb.pathes[name] = path
	sdb.wrappers[name] = wrapper
	delete(sdb.bareDbs, name)
	delete(sdb.queuedDrops, name)
	return wrapper
}

func (sdb *SuperDb) GetDb(name string) kvdb.KeyValueStore {
	if db := sdb.wrappers[name]; db != nil {
		return db
	}
	return sdb.registerNew(name, filepath.Join(sdb.datadir, name+"-ldb"))
}

func (sdb *SuperDb) GetLastDb(prefix string) kvdb.KeyValueStore {
	options := make(map[string]int64)
	for name := range sdb.wrappers {
		if strings.HasPrefix(name, prefix) {
			s := strings.Split(name, "-")
			if len(s) < 2 {
				continue
			}
			indexStr := s[len(s)-2]
			index, err := strconv.ParseInt(indexStr, 10, 64)
			if err != nil {
				continue
			}
			options[name] = index
		}
	}

	maxIndexName := ""
	maxIndex := int64(math.MinInt64)
	for name, index := range options {
		if index > maxIndex {
			maxIndex = index
			maxIndexName = name
		}
	}

	return sdb.GetDb(maxIndexName)
}

func (sdb *SuperDb) DropDb(name string) {
	if db := sdb.bareDbs[name]; db == nil {
		// this DB wasn't flushed, just erase it from RAM then, and that's it
		sdb.erase(name)
		return
	}
	sdb.queuedDrops[name] = struct{}{}
}

func (sdb *SuperDb) erase(name string) {
	delete(sdb.wrappers, name)
	delete(sdb.pathes, name)
	delete(sdb.bareDbs, name)
	delete(sdb.queuedDrops, name)
}

func (sdb *SuperDb) Flush(id hash.Event) error {
	key := []byte("mark")

	// drop old DBs
	for name := range sdb.queuedDrops {
		db := sdb.bareDbs[name]
		if db != nil {
			db.Drop()
		}
		sdb.erase(name)
	}

	// create new DBs, which were not dropped
	for name, wrapper := range sdb.wrappers {
		if db := sdb.bareDbs[name]; db == nil {
			db, err := openDb(sdb.pathes[name])
			if err != nil {
				// TODO log
				return nil
			}
			sdb.bareDbs[name] = db
			wrapper.SetUnderlyingDB(db)
		}
	}

	// write dirty flags
	for _, db := range sdb.bareDbs {
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

	// flush data
	for _, wrapper := range sdb.wrappers {
		wrapper.Flush()
	}

	// write clean flags
	for _, wrapper := range sdb.wrappers {
		err := wrapper.Put(key, id.Bytes())
		if err != nil {
			return err
		}
		wrapper.Flush()
	}

	sdb.prevFlushTime = time.Now()
	return nil
}

func (sdb *SuperDb) FlushIfNeeded(id hash.Event) bool {
	if time.Since(sdb.prevFlushTime) > 10*time.Minute {
		sdb.Flush(id)
		return true
	}

	totalNotFlushed := 0
	for _, db := range sdb.wrappers {
		totalNotFlushed += db.NotFlushedSizeEst()
	}

	if totalNotFlushed > 100*1024*1024 {
		sdb.Flush(id)
		return true
	}
	return false
}

// call on startup, after all dbs are registered
func (sdb *SuperDb) CheckDbsSynced() error {
	key := []byte("mark")
	var prevId *hash.Event
	for _, db := range sdb.bareDbs {
		mark, err := db.Get(key)
		if err != nil {
			return err
		}
		if bytes.HasPrefix(mark, []byte("dirty")) {
			return errors.New("dirty")
		}
		eventId := hash.BytesToEvent(mark)
		if prevId == nil {
			prevId = &eventId
		}
		if eventId != *prevId {
			return errors.New("not synced")
		}
	}
	return nil
}

func openDb(path string) (
	db kvdb.KeyValueStore,
	err error,
) {
	err = os.MkdirAll(path, 0600)
	if err != nil {
		return
	}

	var stopWatcher func()

	onClose := func() error {
		if stopWatcher != nil {
			stopWatcher()
		}
		return nil
	}
	onDrop := func() error {
		return os.RemoveAll(path)
	}

	db, err = leveldb.New(path, 16, 0, "", onClose, onDrop)
	if err != nil {
		panic(fmt.Sprintf("can't create temporary database: %v", err))
	}

	// TODO: dir watcher instead of file watcher needed.
	//stopWatcher = metrics.StartFileWatcher(name+"_db_file_size", f)

	return
}
