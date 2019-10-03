package flushable

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
	"sync"
	"time"

	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/leveldb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
)

type SyncedPool struct {
	pathes   map[string]string
	wrappers map[string]*Flushable
	bareDbs  map[string]kvdb.KeyValueStore

	queuedDrops map[string]struct{}

	prevFlushTime time.Time

	datadir string

	mutex sync.Mutex
}

func NewSyncedPool(datadir string) *SyncedPool {
	dirs, err := ioutil.ReadDir(datadir)
	if err != nil {
		println(err.Error())
		return nil
	}

	p := &SyncedPool{
		pathes:   make(map[string]string),
		wrappers: make(map[string]*Flushable),
		bareDbs:  make(map[string]kvdb.KeyValueStore),

		queuedDrops: make(map[string]struct{}),
		datadir:     datadir,
	}

	for _, f := range dirs {
		dirname := f.Name()
		if f.IsDir() && strings.HasSuffix(dirname, "-ldb") {
			name := strings.TrimSuffix(dirname, "-ldb")
			path := filepath.Join(datadir, dirname)
			p.registerExisting(name, path)
		}
	}
	return p
}

func (p *SyncedPool) registerExisting(name, path string) kvdb.KeyValueStore {
	db, err := openDb(path)
	if err != nil {
		println(err.Error())
		return nil
	}
	wrapper := New(db)

	p.pathes[name] = path
	p.bareDbs[name] = db
	p.wrappers[name] = wrapper
	delete(p.queuedDrops, name)
	return wrapper
}

func (p *SyncedPool) registerNew(name, path string) kvdb.KeyValueStore {
	wrapper := New(memorydb.New())

	p.pathes[name] = path
	p.wrappers[name] = wrapper
	delete(p.bareDbs, name)
	delete(p.queuedDrops, name)
	return wrapper
}

func (p *SyncedPool) GetDbByIndex(prefix string, index int64) kvdb.KeyValueStore {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.getDb(fmt.Sprintf("%s-%d", prefix, index))
}

func (p *SyncedPool) GetDb(name string) kvdb.KeyValueStore {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.getDb(name)
}

func (p *SyncedPool) getDb(name string) kvdb.KeyValueStore {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if wrapper := p.wrappers[name]; wrapper != nil {
		return wrapper
	}
	return p.registerNew(name, filepath.Join(p.datadir, name+"-ldb"))
}

func (p *SyncedPool) GetLastDb(prefix string) kvdb.KeyValueStore {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	options := make(map[string]int64)
	for name := range p.wrappers {
		if strings.HasPrefix(name, prefix) {
			s := strings.Split(name, "-")
			if len(s) < 2 {
				println(name, "name without index")
				continue
			}
			indexStr := s[len(s)-1]
			index, err := strconv.ParseInt(indexStr, 10, 64)
			if err != nil {
				println(err.Error())
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

	return p.getDb(maxIndexName)
}

func (p *SyncedPool) DropDb(name string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if db := p.bareDbs[name]; db == nil {
		// this DB wasn't flushed, just erase it from RAM then, and that's it
		p.erase(name)
		return
	}
	p.queuedDrops[name] = struct{}{}
}

func (p *SyncedPool) erase(name string) {
	delete(p.wrappers, name)
	delete(p.pathes, name)
	delete(p.bareDbs, name)
	delete(p.queuedDrops, name)
}

func (p *SyncedPool) Flush(id []byte) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	return p.flush(id)
}

func (p *SyncedPool) flush(id []byte) error {
	key := []byte("flag")

	// drop old DBs
	for name := range p.queuedDrops {
		db := p.bareDbs[name]
		if db != nil {
			db.Close()
			db.Drop()
		}
		p.erase(name)
	}

	// create new DBs, which were not dropped
	for name, wrapper := range p.wrappers {
		if db := p.bareDbs[name]; db == nil {
			db, err := openDb(p.pathes[name])
			if err != nil {
				println(err.Error())
				return nil
			}
			p.bareDbs[name] = db
			wrapper.SetUnderlyingDB(db)
		}
	}

	// write dirty flags
	for _, db := range p.bareDbs {
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
		marker.Write(id)
		err = db.Put(key, marker.Bytes())
		if err != nil {
			return err
		}
	}

	// flush data
	for _, wrapper := range p.wrappers {
		wrapper.Flush()
	}

	// write clean flags
	for _, wrapper := range p.wrappers {
		err := wrapper.Put(key, id)
		if err != nil {
			return err
		}
		wrapper.Flush()
	}

	p.prevFlushTime = time.Now()
	return nil
}

func (p *SyncedPool) FlushIfNeeded(id []byte) bool {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if time.Since(p.prevFlushTime) > 10*time.Minute {
		p.Flush(id)
		return true
	}

	totalNotFlushed := 0
	for _, db := range p.wrappers {
		totalNotFlushed += db.NotFlushedSizeEst()
	}

	if totalNotFlushed > 100*1024*1024 {
		p.Flush(id)
		return true
	}
	return false
}

// call on startup, after all dbs are registered
func (p *SyncedPool) CheckDbsSynced() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	key := []byte("flag")
	var prevId *[]byte
	for _, db := range p.bareDbs {
		mark, err := db.Get(key)
		if err != nil {
			return err
		}
		if bytes.HasPrefix(mark, []byte("dirty")) {
			return errors.New("dirty")
		}
		if prevId == nil {
			prevId = &mark
		}
		if bytes.Compare(mark, *prevId) != 0 {
			return errors.New("not synced")
		}
	}
	return nil
}

func (p *SyncedPool) CloseAll() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, db := range p.bareDbs {
		if err := db.Close(); err != nil {
			return err
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
