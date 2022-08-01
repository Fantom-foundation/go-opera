package integration

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/Fantom-foundation/lachesis-base/hash"
	"github.com/Fantom-foundation/lachesis-base/inter/dag"
	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/flaggedproducer"
	"github.com/Fantom-foundation/lachesis-base/kvdb/flushable"
	"github.com/Fantom-foundation/lachesis-base/kvdb/leveldb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/multidb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/pebble"
	"github.com/Fantom-foundation/lachesis-base/utils/fmtfilter"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/log"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/Fantom-foundation/go-opera/gossip"
	"github.com/Fantom-foundation/go-opera/utils/dbutil/asyncflushproducer"
)

type DBsConfig struct {
	Routing      RoutingConfig
	RuntimeCache DBsCacheConfig
	GenesisCache DBsCacheConfig
}

type DBCacheConfig struct {
	Cache   uint64
	Fdlimit uint64
}

type DBsCacheConfig struct {
	Table map[string]DBCacheConfig
}

func DefaultDBsConfig(scale func(uint64) uint64, fdlimit uint64) DBsConfig {
	return DBsConfig{
		Routing:      DefaultRoutingConfig(),
		RuntimeCache: DefaultRuntimeDBsCacheConfig(scale, fdlimit),
		GenesisCache: DefaultGenesisDBsCacheConfig(scale, fdlimit),
	}
}

func DefaultRuntimeDBsCacheConfig(scale func(uint64) uint64, fdlimit uint64) DBsCacheConfig {
	return DBsCacheConfig{
		Table: map[string]DBCacheConfig{
			"evm-data": {
				Cache:   scale(242 * opt.MiB),
				Fdlimit: fdlimit*242/700 + 1,
			},
			"evm-logs": {
				Cache:   scale(110 * opt.MiB),
				Fdlimit: fdlimit*110/700 + 1,
			},
			"main": {
				Cache:   scale(186 * opt.MiB),
				Fdlimit: fdlimit*186/700 + 1,
			},
			"events": {
				Cache:   scale(87 * opt.MiB),
				Fdlimit: fdlimit*87/700 + 1,
			},
			"epoch-%d": {
				Cache:   scale(75 * opt.MiB),
				Fdlimit: fdlimit*75/700 + 1,
			},
			"": {
				Cache:   64 * opt.MiB,
				Fdlimit: fdlimit/100 + 1,
			},
		},
	}
}

func DefaultGenesisDBsCacheConfig(scale func(uint64) uint64, fdlimit uint64) DBsCacheConfig {
	return DBsCacheConfig{
		Table: map[string]DBCacheConfig{
			"main": {
				Cache:   scale(1024 * opt.MiB),
				Fdlimit: fdlimit*1024/3072 + 1,
			},
			"evm-data": {
				Cache:   scale(1024 * opt.MiB),
				Fdlimit: fdlimit*1024/3072 + 1,
			},
			"evm-logs": {
				Cache:   scale(1024 * opt.MiB),
				Fdlimit: fdlimit*1024/3072 + 1,
			},
			"events": {
				Cache:   scale(1 * opt.MiB),
				Fdlimit: fdlimit*1/3072 + 1,
			},
			"epoch-%d": {
				Cache:   scale(1 * opt.MiB),
				Fdlimit: fdlimit*1/3072 + 1,
			},
			"": {
				Cache:   16 * opt.MiB,
				Fdlimit: fdlimit/100 + 1,
			},
		},
	}
}

func SupportedDBs(chaindataDir string, cfg DBsCacheConfig) (map[multidb.TypeName]kvdb.IterableDBProducer, map[multidb.TypeName]kvdb.FullDBProducer) {
	if chaindataDir == "inmemory" || chaindataDir == "" {
		chaindataDir, _ = ioutil.TempDir("", "opera-tmp")
	}
	cacher, err := dbCacheFdlimit(cfg)
	if err != nil {
		utils.Fatalf("Failed to create DB cacher: %v", err)
	}

	leveldbFsh := leveldb.NewProducer(path.Join(chaindataDir, "leveldb-fsh"), cacher)
	leveldbFlg := leveldb.NewProducer(path.Join(chaindataDir, "leveldb-flg"), cacher)
	leveldbDrc := leveldb.NewProducer(path.Join(chaindataDir, "leveldb-drc"), cacher)
	pebbleFsh := pebble.NewProducer(path.Join(chaindataDir, "pebble-fsh"), cacher)
	pebbleFlg := pebble.NewProducer(path.Join(chaindataDir, "pebble-flg"), cacher)
	pebbleDrc := pebble.NewProducer(path.Join(chaindataDir, "pebble-drc"), cacher)

	return map[multidb.TypeName]kvdb.IterableDBProducer{
			"leveldb-fsh": leveldbFsh,
			"leveldb-flg": leveldbFlg,
			"leveldb-drc": leveldbDrc,
			"pebble-fsh":  pebbleFsh,
			"pebble-flg":  pebbleFlg,
			"pebble-drc":  pebbleDrc,
		}, map[multidb.TypeName]kvdb.FullDBProducer{
			"leveldb-fsh": flushable.NewSyncedPool(leveldbFsh, FlushIDKey),
			"leveldb-flg": flaggedproducer.Wrap(leveldbFlg, FlushIDKey),
			"leveldb-drc": &DummyScopedProducer{leveldbDrc},
			"pebble-fsh":  asyncflushproducer.Wrap(flushable.NewSyncedPool(pebbleFsh, FlushIDKey), 200000),
			"pebble-flg":  flaggedproducer.Wrap(pebbleFlg, FlushIDKey),
			"pebble-drc":  &DummyScopedProducer{pebbleDrc},
		}
}

func dbCacheFdlimit(cfg DBsCacheConfig) (func(string) (int, int), error) {
	fmts := make([]func(req string) (string, error), 0, len(cfg.Table))
	fmtsCaches := make([]DBCacheConfig, 0, len(cfg.Table))
	exactTable := make(map[string]DBCacheConfig, len(cfg.Table))
	// build scanf filters
	for name, cache := range cfg.Table {
		if !strings.ContainsRune(name, '%') {
			exactTable[name] = cache
		} else {
			fn, err := fmtfilter.CompileFilter(name, name)
			if err != nil {
				return nil, err
			}
			fmts = append(fmts, fn)
			fmtsCaches = append(fmtsCaches, cache)
		}
	}
	return func(name string) (int, int) {
		// try exact match
		if cache, ok := cfg.Table[name]; ok {
			return int(cache.Cache), int(cache.Fdlimit)
		}
		// try regexp
		for i, fn := range fmts {
			if _, err := fn(name); err == nil {
				return int(fmtsCaches[i].Cache), int(fmtsCaches[i].Fdlimit)
			}
		}
		// default
		return int(cfg.Table[""].Cache), int(cfg.Table[""].Fdlimit)
	}, nil
}

func isEmpty(dir string) bool {
	f, err := os.Open(dir)
	if err != nil {
		return true
	}
	defer f.Close()
	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true
	}
	return false
}

func dropAllDBs(chaindataDir string) {
	_ = os.RemoveAll(chaindataDir)
}

func dropAllDBsIfInterrupted(chaindataDir string) {
	if isInterrupted(chaindataDir) {
		log.Info("Restarting genesis processing")
		dropAllDBs(chaindataDir)
	}
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

func MakeDBDirs(chaindataDir string) {
	dbs, _ := SupportedDBs(chaindataDir, DBsCacheConfig{})
	for typ := range dbs {
		if err := os.MkdirAll(path.Join(chaindataDir, string(typ)), 0700); err != nil {
			utils.Fatalf("Failed to create chaindata/leveldb directory: %v", err)
		}
	}
}

type DummyScopedProducer struct {
	kvdb.IterableDBProducer
}

func (d DummyScopedProducer) NotFlushedSizeEst() int {
	return 0
}

func (d DummyScopedProducer) Flush(_ []byte) error {
	return nil
}

func (d DummyScopedProducer) Initialize(_ []string, flushID []byte) ([]byte, error) {
	return flushID, nil
}

func (d DummyScopedProducer) Close() error {
	return nil
}
