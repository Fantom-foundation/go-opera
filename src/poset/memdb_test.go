package poset

import (
	"math/rand"

	"github.com/ethereum/go-ethereum/log"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/fallible"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/memorydb"
)

type fakeFS struct {
	Namespace string
	Files     map[string]kvdb.KeyValueStore
}

var fakeFSs = make(map[string]*fakeFS)

func newFakeFS(namespace string) *fakeFS {
	if fs, ok := fakeFSs[namespace]; ok {
		return fs
	}

	fs := &fakeFS{
		Namespace: namespace,
		Files:     make(map[string]kvdb.KeyValueStore),
	}
	fakeFSs[namespace] = fs
	return fs
}

func uniqNamespace() string {
	return hash.FakeHash(rand.Int63()).Hex()
}

func (fs *fakeFS) OpenFakeDB(name string) kvdb.KeyValueStore {
	if db, ok := fs.Files[name]; ok {
		mem := db.(*fallible.Fallible).Underlying.(*memorydb.Database)
		log.Debug("open fake-DB", "db-name", name, "namespace", fs.Namespace, "len", mem.Len())

		return db
	}

	mem := memorydb.New()
	log.Debug("make fake-DB", "db-name", name, "namespace", fs.Namespace, "len", "0")

	db := fallible.Wrap(mem, nil,
		func() error { // on drop
			log.Debug("drop fake-DB", "db-name", name, "namespace", fs.Namespace, "mem", mem.Len())
			delete(fs.Files, name)
			return nil
		},
	)
	db.SetWriteCount(enough)

	fs.Files[name] = db

	return db
}
