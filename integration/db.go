package integration

import (
	"errors"
	"strings"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/flushable"
	"github.com/Fantom-foundation/lachesis-base/kvdb/leveldb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/memorydb"
	"github.com/Fantom-foundation/lachesis-base/utils/cachescale"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/Fantom-foundation/go-opera/gossip"
)

func DBProducer(chaindataDir string, scale cachescale.Func) kvdb.IterableDBProducer {
	if chaindataDir == "inmemory" || chaindataDir == "" {
		return memorydb.NewProducer("")
	}

	return leveldb.NewProducer(chaindataDir, func(name string) int {
		return dbCacheSize(name, scale.I)
	})
}

func CheckDBList(names []string) error {
	if len(names) == 0 {
		return nil
	}
	namesMap := make(map[string]bool)
	for _, name := range names {
		namesMap[name] = true
	}
	if !namesMap["gossip"] {
		return errors.New("gossip DB is not found")
	}
	if !namesMap["lachesis"] {
		return errors.New("lachesis DB is not found")
	}

	return nil
}

func dbCacheSize(name string, scale func(int) int) int {
	if name == "gossip" {
		return scale(128 * opt.MiB)
	}
	if name == "lachesis" {
		return scale(4 * opt.MiB)
	}
	if strings.HasPrefix(name, "lachesis-") {
		return scale(8 * opt.MiB)
	}
	if strings.HasPrefix(name, "gossip-") {
		return scale(8 * opt.MiB)
	}
	return scale(2 * opt.MiB)
}

func dropAllDBs(producer kvdb.IterableDBProducer) {
	names := producer.Names()
	for _, name := range names {
		db, err := producer.OpenDB(name)
		if err != nil {
			continue
		}
		_ = db.Close()
		db.Drop()
	}
}

func dropAllDBsIfInterrupted(rawProducer kvdb.IterableDBProducer) {
	names := rawProducer.Names()
	if len(names) == 0 {
		return
	}
	// if flushID is not written, then previous genesis processing attempt was interrupted
	for _, name := range names {
		db, err := rawProducer.OpenDB(name)
		if err != nil {
			return
		}
		flushID, err := db.Get(FlushIDKey)
		_ = db.Close()
		if flushID != nil || err != nil {
			return
		}
	}
	dropAllDBs(rawProducer)
}

type GossipStoreAdapter struct {
	*gossip.Store
}

func (g *GossipStoreAdapter) GetEvent(id hash.Event) dag.Event {
	e := g.Store.GetEvent(id)
	if e == nil {
		return nil
	}
	return e
}

type dummyFlushableProducer struct {
	kvdb.IterableDBProducer
	opened map[string]kvdb.DropableStore
}

func NewDummyFlushableProducer(dbs kvdb.IterableDBProducer) kvdb.FlushableDBProducer {
	return &dummyFlushableProducer{
		IterableDBProducer: dbs,
		opened:             make(map[string]kvdb.DropableStore),
	}
}

func (p *dummyFlushableProducer) NotFlushedSizeEst() int {
	return 0
}

func (p *dummyFlushableProducer) OpenDB(name string) (kvdb.DropableStore, error) {
	if db, ok := p.opened[name]; ok {
		return db, nil
	}

	db, err := p.IterableDBProducer.OpenDB(name)
	if err != nil {
		return nil, err
	}

	p.opened[name] = db
	return db, nil
}

func (p *dummyFlushableProducer) Flush(id []byte) error {
	for _, name := range p.Names() {
		db, err := p.OpenDB(name)
		if err != nil {
			// skip dropped
			continue
		}

		err = db.Put(FlushIDKey, append([]byte{flushable.CleanPrefix}, id...))
		if err != nil {
			return err
		}
	}
	return nil
}
