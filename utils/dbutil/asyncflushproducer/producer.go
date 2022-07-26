package asyncflushproducer

import (
	"errors"
	"sync"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/ethereum/go-ethereum/metrics"
)

type Producer struct {
	kvdb.FullDBProducer
	mu    sync.Mutex
	dbs   map[string]*store
	stats metrics.Meter

	threshold uint64
}

func Wrap(backend kvdb.FullDBProducer, threshold uint64) *Producer {
	return &Producer{
		stats:          metrics.NewMeterForced(),
		FullDBProducer: backend,
		dbs:            make(map[string]*store),
		threshold:      threshold,
	}
}

func (f *Producer) OpenDB(name string) (kvdb.Store, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	// open existing DB
	openedDB := f.dbs[name]
	if openedDB != nil {
		return openedDB, nil
	}
	// create new DB
	db, err := f.FullDBProducer.OpenDB(name)
	if err != nil {
		return nil, err
	}
	if f.dbs[name] != nil {
		return nil, errors.New("already opened")
	}
	wrapped := &store{
		Store: db,
		CloseFn: func() error {
			f.mu.Lock()
			delete(f.dbs, name)
			f.mu.Unlock()
			return db.Close()
		},
	}
	f.dbs[name] = wrapped
	return wrapped, nil
}

func (f *Producer) Flush(id []byte) error {
	f.stats.Mark(int64(f.FullDBProducer.NotFlushedSizeEst()))

	err := f.FullDBProducer.Flush(id)
	if err != nil {
		return err
	}

	// trigger flushing data to disk if throughput is below a threshold
	if uint64(f.stats.Rate1()) <= f.threshold {
		go func() {
			f.mu.Lock()
			defer f.mu.Unlock()
			for _, db := range f.dbs {
				_, _ = db.Stat("async_flush")
			}
		}()
	}

	return nil
}
