package main

import (
	"os"
	"path/filepath"

	"go.etcd.io/bbolt"

	"github.com/Fantom-foundation/go-lachesis/src/gossip"
	"github.com/Fantom-foundation/go-lachesis/src/internal"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/metrics"
	"github.com/Fantom-foundation/go-lachesis/src/poset"
)

func makeStorages(makeDb internal.DbProducer) (*gossip.Store, *poset.Store) {
	db := makeDb("lachesis")

	g := db.NewTable([]byte("g_"))
	p := db.NewTable([]byte("p_"))

	return gossip.NewStore(g, makeDb),
		poset.NewStore(p, makeDb)
}

func dbProducer(dbdir string) internal.DbProducer {
	if dbdir == "inmemory" {
		return func(name string) kvdb.Database {
			return kvdb.NewMemDatabase()
		}
	}

	return func(name string) kvdb.Database {
		bdb, close, drop, err := openDb(dbdir, name)
		if err != nil {
			panic(err)
		}

		return kvdb.NewBoltDatabase(bdb, close, drop)
	}
}

func openDb(dir, name string) (
	db *bbolt.DB,
	close, drop func() error,
	err error,
) {
	err = os.MkdirAll(dir, 0600)
	if err != nil {
		return
	}

	f := filepath.Join(dir, name+".bolt")
	db, err = bbolt.Open(f, 0600, nil)
	if err != nil {
		return
	}

	stopWatcher := metrics.StartFileWatcher(name+"_db_file_size", f)

	close = func() error {
		stopWatcher()
		return db.Close()
	}

	drop = func() error {
		return os.Remove(f)
	}

	return
}
