package memorydb

import (
	"github.com/Fantom-foundation/go-lachesis/kvdb"
)

type Mod func(kvdb.KeyValueStore) kvdb.KeyValueStore

type producer struct {
	fs   *fakeFS
	mods []Mod
}

// NewProducer of memory db.
func NewProducer(namespace string, mods ...Mod) kvdb.DbProducer {
	return &producer{
		fs:   newFakeFS(namespace),
		mods: mods,
	}
}

// Names of existing databases.
func (p *producer) Names() []string {
	return p.fs.ListFakeDB()
}

// OpenDb or create db with name.
func (p *producer) OpenDb(name string) kvdb.KeyValueStore {
	db := p.fs.OpenFakeDB(name)

	for _, mod := range p.mods {
		db = mod(db)
	}

	return db
}
