package flushable

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-lachesis/kvdb"
)

type SyncedPool struct {
	producer kvdb.DbProducer

	wrappers    map[string]*Flushable
	queuedDrops map[string]struct{}

	prevFlushTime time.Time

	sync.Mutex
}

func NewSyncedPool(producer kvdb.DbProducer) *SyncedPool {
	if producer == nil {
		panic("nil producer")
	}

	p := &SyncedPool{
		producer:    producer,
		wrappers:    make(map[string]*Flushable),
		queuedDrops: make(map[string]struct{}),
	}

	for _, name := range producer.Names() {
		open, drop := p.callbacks(name)
		p.wrappers[name] = New(open, drop)
	}

	if err := p.checkDbsSynced(); err != nil {
		log.Crit("Databases aren't sync.")
	}

	return p
}

func (p *SyncedPool) callbacks(name string) (
	onOpen func() kvdb.KeyValueStore,
	onDrop func(),
) {
	onOpen = func() kvdb.KeyValueStore {
		return p.producer.OpenDb(name)
	}

	onDrop = func() {
		p.dropDb(name)
	}

	return
}

func (p *SyncedPool) GetDbByIndex(prefix string, index int64) kvdb.KeyValueStore {
	p.Lock()
	defer p.Unlock()

	return p.getDb(fmt.Sprintf("%s-%d", prefix, index))
}

func (p *SyncedPool) GetDb(name string) kvdb.KeyValueStore {
	p.Lock()
	defer p.Unlock()

	return p.getDb(name)
}

func (p *SyncedPool) getDb(name string) kvdb.KeyValueStore {
	wrapper := p.wrappers[name]
	if wrapper != nil {
		return wrapper
	}

	open, drop := p.callbacks(name)
	wrapper = New(open, drop)

	p.wrappers[name] = wrapper

	return wrapper
}

func (p *SyncedPool) GetLastDb(prefix string) kvdb.KeyValueStore {
	p.Lock()
	defer p.Unlock()

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

func (p *SyncedPool) dropDb(name string) {
	p.Lock()
	defer p.Unlock()

	p.queuedDrops[name] = struct{}{}
}

func (p *SyncedPool) erase(name string) {
	delete(p.wrappers, name)
	delete(p.queuedDrops, name)
}

func (p *SyncedPool) Flush(id []byte) error {
	p.Lock()
	defer p.Unlock()

	return p.flush(id)
}

func (p *SyncedPool) flush(id []byte) error {
	key := []byte("flag")

	// drop old DBs
	for name := range p.queuedDrops {
		w := p.wrappers[name]
		p.erase(name)
		if w == nil {
			continue
		}
		db := w.UnderlyingDb()
		if db == nil {
			continue
		}
		err := db.Close()
		if err != nil {
			return err
		}
		db.Drop()
	}

	// write dirty flags
	for _, w := range p.wrappers {
		db := w.UnderlyingDb()

		prev, err := db.Get(key)
		if err != nil {
			return err
		}
		if prev == nil {
			prev = []byte("initial")
		}

		marker := bytes.NewBuffer(nil)
		marker.Write([]byte("dirty"))
		marker.Write(prev)
		marker.Write(id)
		err = db.Put(key, marker.Bytes())
		if err != nil {
			return err
		}
	}

	// flush data
	for _, wrapper := range p.wrappers {
		err := wrapper.Flush()
		if err != nil {
			return err
		}
	}

	// write clean flags
	for _, w := range p.wrappers {
		db := w.UnderlyingDb()
		err := db.Put(key, id)
		if err != nil {
			return err
		}
	}

	p.prevFlushTime = time.Now()
	return nil
}

func (p *SyncedPool) FlushIfNeeded(id []byte) bool {
	p.Lock()
	defer p.Unlock()

	if time.Since(p.prevFlushTime) > 10*time.Minute {
		p.flush(id)
		return true
	}

	totalNotFlushed := 0
	for _, db := range p.wrappers {
		totalNotFlushed += db.NotFlushedSizeEst()
	}

	if totalNotFlushed > 100*1024*1024 {
		p.flush(id)
		return true
	}
	return false
}

// checkDbsSynced on startup, after all dbs are registered.
func (p *SyncedPool) checkDbsSynced() error {
	p.Lock()
	defer p.Unlock()

	key := []byte("flag")
	var prevId *[]byte
	for _, w := range p.wrappers {
		db := w.UnderlyingDb()

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
	p.Lock()
	defer p.Unlock()

	for _, db := range p.wrappers {
		if err := db.Close(); err != nil {
			return err
		}
	}
	return nil
}
