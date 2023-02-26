package threads

import (
	"github.com/Fantom-foundation/lachesis-base/kvdb"

	"github.com/Fantom-foundation/go-opera/logger"
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
		release func(count int)
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
	got, release := GlobalPool.Lock(1)
	if got < 1 {
		notifier.Log.Warn("Too much db iterators")
	}

	return &limitedIterator{
		Iterator: s.Store.NewIterator(prefix, start),
		release:  release,
	}
}

func (it *limitedIterator) Release() {
	it.Iterator.Release()
	it.release(1)
}
