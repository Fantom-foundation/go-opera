package threads

import (
	"fmt"
	"time"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	// "github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-opera/logger"
)

const (
	newIteratorTimeout = 3 * time.Second
)

type (
	limitedDbProducer struct {
		kvdb.DBProducer
	}

	limitedFullDbProducer struct {
		kvdb.FullDBProducer
	}

	limitedStore struct {
		kvdb.Store
	}

	limitedIterator struct {
		kvdb.Iterator
		release func()
	}
)

func LimitedDBProducer(dbs kvdb.DBProducer) kvdb.DBProducer {
	return &limitedDbProducer{dbs}
}

func LimitedFullDBProducer(dbs kvdb.FullDBProducer) kvdb.FullDBProducer {
	return &limitedFullDbProducer{dbs}
}

func (p *limitedDbProducer) OpenDB(name string) (kvdb.Store, error) {
	s, err := p.DBProducer.OpenDB(name)
	return &limitedStore{s}, err
}

func (p *limitedFullDbProducer) OpenDB(name string) (kvdb.Store, error) {
	s, err := p.FullDBProducer.OpenDB(name)
	return &limitedStore{s}, err
}

var notifier = logger.New("threads-pool")

func (s *limitedStore) NewIterator(prefix []byte, start []byte) kvdb.Iterator {
	timeout := time.After(newIteratorTimeout)

	got, release := GlobalPool.Lock(1)
	for ; got < 1; got, release = GlobalPool.Lock(1) {
		// wait for free pool item
		release()
		select {
		case <-time.After(time.Millisecond):
			continue
		case <-timeout:
			notifier.Log.Warn("No free threads to open db iterator", "timeout", fmt.Sprintf("%ds", newIteratorTimeout/time.Second))
		}
	}

	return &limitedIterator{
		Iterator: s.Store.NewIterator(prefix, start),
		release:  release,
	}
}

func (it *limitedIterator) Release() {
	it.Iterator.Release()
	it.release()
}
