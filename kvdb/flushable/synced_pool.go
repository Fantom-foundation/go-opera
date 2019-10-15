package flushable

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/status-im/keycard-go/hexutils"

	"github.com/Fantom-foundation/go-lachesis/kvdb"
)

type SyncedPool struct {
	producer kvdb.DbProducer

	wrappers    map[string]*LazyFlushable
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
		wrappers:    make(map[string]*LazyFlushable),
		queuedDrops: make(map[string]struct{}),
	}

	for _, name := range producer.Names() {
		open, drop := p.callbacks(name)
		p.wrappers[name] = NewLazy(open, drop)
	}

	if err := p.checkDbsSynced(); err != nil {
		log.Crit("Databases are corrupted, which is possible after a crash or disk failure.", "err", err)
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

func (p *SyncedPool) dropDb(name string) {
	p.Lock()
	defer p.Unlock()

	p.queuedDrops[name] = struct{}{}
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
	wrapper = NewLazy(open, drop)

	p.wrappers[name] = wrapper

	return wrapper
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
		delete(p.wrappers, name)
		if w == nil {
			continue
		}
		db := w.underlying
		if db == nil {
			continue
		}
		// db.Close() is called inside wrapper.Close()
		db.Drop()
	}
	p.queuedDrops = make(map[string]struct{})

	// write dirty flags
	for _, w := range p.wrappers {
		db := w.InitUnderlyingDb()

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
		db := w.InitUnderlyingDb()
		err := db.Put(key, id)
		if err != nil {
			return err
		}
	}

	p.prevFlushTime = time.Now()
	return nil
}

func (p *SyncedPool) FlushIfNeeded(id []byte) (bool, error) {
	p.Lock()
	defer p.Unlock()

	if time.Since(p.prevFlushTime) > 10*time.Minute {
		return true, p.flush(id)
	}

	totalNotFlushed := 0
	for _, db := range p.wrappers {
		totalNotFlushed += db.NotFlushedSizeEst()
	}

	if totalNotFlushed > 10*1024*1024 {
		return true, p.flush(id)
	}
	return false, nil
}

// checkDbsSynced on startup, after all dbs are registered.
func (p *SyncedPool) checkDbsSynced() error {
	p.Lock()
	defer p.Unlock()

	var (
		key    = []byte("flag")
		prevId *[]byte
		descrs []string
		list   = func() string {
			return strings.Join(descrs, ",\n")
		}
	)
	for name, w := range p.wrappers {
		db := w.InitUnderlyingDb()

		mark, err := db.Get(key)
		if err != nil {
			return err
		}
		descrs = append(descrs, fmt.Sprintf("%s: %s", name, hexutils.BytesToHex(mark)))

		if bytes.HasPrefix(mark, []byte("dirty")) {
			return fmt.Errorf("dirty state: %s", list())
		}
		if prevId == nil {
			prevId = &mark
		}
		if !bytes.Equal(mark, *prevId) {
			return fmt.Errorf("not synced: %s", list())
		}
	}
	return nil
}
