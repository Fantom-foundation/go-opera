package threads

import (
	"time"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
)

type limitedProducer struct {
	kvdb.FullDBProducer
}

type limitedStore struct {
	kvdb.Store
}

type limitedIterator struct {
	kvdb.Iterator
	release func()
}

func Limited(dbs kvdb.FullDBProducer) kvdb.FullDBProducer {
	return &limitedProducer{dbs}
}

func (p *limitedProducer) OpenDB(name string) (kvdb.Store, error) {
	s, err := p.FullDBProducer.OpenDB(name)
	return &limitedStore{s}, err
}

func (s *limitedStore) NewIterator(prefix []byte, start []byte) kvdb.Iterator {
	got, release := globalPool.Lock(1)
	for ; got < 1; got, release = globalPool.Lock(1) {
		// wait for free pool item
		release()
		<-time.After(time.Millisecond)
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
