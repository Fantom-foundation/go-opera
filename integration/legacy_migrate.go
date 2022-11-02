package integration

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/batched"
	"github.com/Fantom-foundation/lachesis-base/kvdb/leveldb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/pebble"
	"github.com/Fantom-foundation/lachesis-base/kvdb/skipkeys"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/Fantom-foundation/go-opera/utils/compactdb"
)

func lastKey(db kvdb.Store) []byte {
	var start []byte
	for {
		for b := 0xff; b >= 0; b-- {
			if !isEmptyDB(table.New(db, append(start, byte(b)))) {
				start = append(start, byte(b))
				break
			}
			if b == 0 {
				return start
			}
		}
	}
}

type transformTask struct {
	openSrc func() kvdb.Store
	openDst func() kvdb.Store
	name    string
	dir     string
	dropSrc bool
}

func transform(m transformTask) error {
	openDst := func() *batched.Store {
		return batched.Wrap(m.openDst())
	}
	openSrc := func() *batched.Store {
		return batched.Wrap(m.openSrc())
	}
	src := openSrc()
	defer func() {
		_ = src.Close()
		if m.dropSrc {
			src.Drop()
		}
	}()
	if isEmptyDB(src) {
		return nil
	}
	dst := openDst()

	const batchKeys = 5000000
	keys := make([][]byte, 0, batchKeys)
	// start from previously written data, if any
	it := src.NewIterator(nil, lastKey(dst))
	defer func() {
		// wrap with func because DBs may be reopened below
		it.Release()
		_ = dst.Close()
	}()
	log.Info("Transforming DB layout", "db", m.name)
	for next := true; next; {
		for len(keys) < batchKeys {
			next = it.Next()
			if !next {
				break
			}
			err := dst.Put(it.Key(), it.Value())
			if err != nil {
				utils.Fatalf("Failed to put: %v", err)
			}
			keys = append(keys, common.CopyBytes(it.Key()))
		}
		err := dst.Flush()
		if err != nil {
			utils.Fatalf("Failed to flush: %v", err)
		}
		freeSpace, err := getFreeDiskSpace(m.dir)
		if err != nil {
			log.Error("Failed to retrieve free disk space", "err", err)
		} else if freeSpace < 20*opt.GiB {
			return errors.New("not enough disk space")
		} else if len(keys) > 0 && freeSpace < 100*opt.GiB {
			log.Warn("Running out of disk space. Trimming source DB records", "space_GB", freeSpace/opt.GiB)
			_, _ = dst.Stat("async_flush")
			// release iterator so that DB could release data
			it.Release()
			// erase data from src
			for _, k := range keys {
				_ = src.Delete(k)
			}
			_ = src.Compact(keys[0], keys[len(keys)-1])
			// reopen source DB too if it doesn't release data
			if freeSpace < 50*opt.GiB {
				_ = src.Close()
				src = openSrc()
			}
			it = src.NewIterator(nil, keys[len(keys)-1])
		}
		keys = keys[:0]
	}
	// compact the new DB
	if err := compactdb.Compact(dst, m.name); err != nil {
		return err
	}
	return nil
}

func mustTransform(m transformTask) {
	err := transform(m)
	if err != nil {
		utils.Fatalf(err.Error())
	}
}

func isEmptyDB(db kvdb.Iteratee) bool {
	it := db.NewIterator(nil, nil)
	defer it.Release()
	return !it.Next()
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

type closebaleTable struct {
	*table.Table
	backend kvdb.Store
}

func (s *closebaleTable) Close() error {
	return s.backend.Close()
}

func (s *closebaleTable) Drop() {
	s.backend.Drop()
}

func newClosableTable(db kvdb.Store, prefix []byte) *closebaleTable {
	return &closebaleTable{
		Table:   table.New(db, prefix),
		backend: db,
	}
}

func translateGossipPrefix(p byte) byte {
	if p == byte('!') {
		return byte('S')
	}
	if p == byte('@') {
		return byte('R')
	}
	if p == byte('#') {
		return byte('Q')
	}
	if p == byte('$') {
		return byte('T')
	}
	if p == byte('%') {
		return byte('J')
	}
	if p == byte('^') {
		return byte('E')
	}
	if p == byte('&') {
		return byte('I')
	}
	if p == byte('*') {
		return byte('G')
	}
	if p == byte('(') {
		return byte('F')
	}
	return p
}

func equalRoutingConfig(a, b RoutingConfig) bool {
	if len(a.Table) != len(b.Table) {
		return false
	}
	for k, v := range a.Table {
		if b.Table[k] != v {
			return false
		}
	}
	return true
}

func migrateLegacyDBs(chaindataDir string, dbs kvdb.FlushableDBProducer, mode string, layout RoutingConfig) error {
	{ // didn't erase the brackets to avoid massive code changes
		// migrate DB layout
		cacheFn, err := dbCacheFdlimit(DBsCacheConfig{
			Table: map[string]DBCacheConfig{
				"": {
					Cache:   1024 * opt.MiB,
					Fdlimit: uint64(utils.MakeDatabaseHandles() / 2),
				},
			},
		})
		if err != nil {
			return err
		}
		var oldDBs kvdb.IterableDBProducer
		var oldDBsType string
		if fileExists(path.Join(chaindataDir, "gossip", "LOG")) {
			oldDBs = leveldb.NewProducer(chaindataDir, cacheFn)
			oldDBsType = "ldb"
		} else {
			oldDBs = pebble.NewProducer(chaindataDir, cacheFn)
			oldDBsType = "pbl"
		}
		openOldDB := func(name string) kvdb.Store {
			db, err := oldDBs.OpenDB(name)
			if err != nil {
				utils.Fatalf("Failed to open %s old DB: %v", name, err)
			}
			return db
		}
		openNewDB := func(name string) kvdb.Store {
			db, err := dbs.OpenDB(name)
			if err != nil {
				utils.Fatalf("Failed to open %s DB: %v", name, err)
			}
			return db
		}

		switch mode {
		case "rebuild":
			// move lachesis, lachesis-%d and gossip-%d DBs
			for _, name := range oldDBs.Names() {
				if strings.HasPrefix(name, "lachesis") || strings.HasPrefix(name, "gossip-") {
					mustTransform(transformTask{
						openSrc: func() kvdb.Store {
							return skipkeys.Wrap(openOldDB(name), MetadataPrefix)
						},
						openDst: func() kvdb.Store {
							return openNewDB(name)
						},
						name: name,
						dir:  chaindataDir,
					})
				}
			}

			// move gossip DB

			// move logs
			mustTransform(transformTask{
				openSrc: func() kvdb.Store {
					return newClosableTable(openOldDB("gossip"), []byte("Lr"))
				},
				openDst: func() kvdb.Store {
					return openNewDB("evm-logs/r")
				},
				name: "gossip/Lr",
				dir:  chaindataDir,
			})
			mustTransform(transformTask{
				openSrc: func() kvdb.Store {
					return newClosableTable(openOldDB("gossip"), []byte("Lt"))
				},
				openDst: func() kvdb.Store {
					return openNewDB("evm-logs/t")
				},
				name: "gossip/Lt",
				dir:  chaindataDir,
			})

			// skip 0 prefix, as it contains flushID
			for b := 1; b <= 0xff; b++ {
				if b == int('L') {
					// logs are already moved above
					continue
				}
				mustTransform(transformTask{
					openSrc: func() kvdb.Store {
						return newClosableTable(openOldDB("gossip"), []byte{byte(b)})
					},
					openDst: func() kvdb.Store {
						if b == int('M') || b == int('r') || b == int('x') || b == int('X') {
							return openNewDB("evm/" + string([]byte{byte(b)}))
						} else {
							return openNewDB("gossip/" + string([]byte{translateGossipPrefix(byte(b))}))
						}
					},
					name:    fmt.Sprintf("gossip/%c", rune(b)),
					dir:     chaindataDir,
					dropSrc: b == 0xff,
				})
			}
		case "reformat":
			if oldDBsType == "ldb" {
				if !equalRoutingConfig(layout, LdbLegacyRoutingConfig()) {
					return errors.New("reformatting DBs: missing --db.preset=legacy-ldb flag")
				}
				err = os.Rename(path.Join(chaindataDir, "gossip"), path.Join(chaindataDir, "leveldb-fsh", "main"))
				if err != nil {
					return err
				}
				for _, name := range oldDBs.Names() {
					if strings.HasPrefix(name, "lachesis") || strings.HasPrefix(name, "gossip-") {
						err = os.Rename(path.Join(chaindataDir, name), path.Join(chaindataDir, "leveldb-fsh", name))
						if err != nil {
							return err
						}
					}
				}
			} else {
				if !equalRoutingConfig(layout, PblLegacyRoutingConfig()) {
					return errors.New("reformatting DBs: missing --db.preset=legacy-pbl flag")
				}
				err = os.Rename(path.Join(chaindataDir, "gossip"), path.Join(chaindataDir, "pebble-fsh", "main"))
				if err != nil {
					return err
				}
				for _, name := range oldDBs.Names() {
					if strings.HasPrefix(name, "lachesis") || strings.HasPrefix(name, "gossip-") {
						err = os.Rename(path.Join(chaindataDir, name), path.Join(chaindataDir, "pebble-fsh", name))
						if err != nil {
							return err
						}
					}
				}
			}
		default:
			return errors.New("missing --db.migration.mode flag")
		}
	}

	return nil
}
