package integration

import (
	"fmt"
	"path"
	"strings"

	"github.com/Fantom-foundation/lachesis-base/kvdb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/batched"
	"github.com/Fantom-foundation/lachesis-base/kvdb/leveldb"
	"github.com/Fantom-foundation/lachesis-base/kvdb/skipkeys"
	"github.com/Fantom-foundation/lachesis-base/kvdb/table"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

func moveDB(src, dst_ kvdb.Store, name, dir string) error {
	dst := batched.Wrap(dst_)
	defer dst.Flush()

	// start from previously written data, if any
	var start []byte
	for b := 0xff; b > 0; b-- {
		if !isEmptyDB(table.New(dst, []byte{byte(b)})) {
			start = []byte{byte(b)}
			break
		}
	}

	const batchKeys = 1000000
	keys := make([][]byte, 0, batchKeys)
	values := make([][]byte, 0, batchKeys)
	it := src.NewIterator(nil, start)
	defer it.Release()
	log.Info("Transforming DB layout", "db", name)
	for next := true; next; {
		for len(keys) < batchKeys {
			next = it.Next()
			if !next {
				break
			}
			keys = append(keys, common.CopyBytes(it.Key()))
			values = append(values, common.CopyBytes(it.Value()))
		}
		for i := 0; i < len(keys); i++ {
			err := dst.Put(keys[i], values[i])
			if err != nil {
				return err
			}
		}
		freeSpace, err := getFreeDiskSpace(dir)
		if err != nil {
			log.Error("Failed to retrieve free disk space", "err", err)
		} else if len(keys) > 0 && freeSpace < 50*opt.GiB {
			log.Warn("Not enough disk space. Trimming source DB records", "space_GB", freeSpace/opt.GiB)
			_ = dst.Flush()
			_, _ = dst.Stat("async_flush")
			for i := 0; i < len(keys); i++ {
				err := src.Delete(keys[i])
				if err != nil {
					return err
				}
			}
			_ = src.Compact(keys[0], keys[len(keys)-1])
		}
		keys = keys[:0]
		values = values[:0]
	}
	return nil
}

func isEmptyDB(db kvdb.Iteratee) bool {
	it := db.NewIterator(nil, nil)
	defer it.Release()
	return !it.Next()
}

func migrateLegacyDBs(chaindataDir string, dbs kvdb.FlushableDBProducer) error {
	if !isEmpty(path.Join(chaindataDir, "gossip")) {
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
		oldDBs := leveldb.NewProducer(chaindataDir, cacheFn)

		// move lachesis DB
		lachesisDB, err := oldDBs.OpenDB("lachesis")
		if err != nil {
			return err
		}
		newDB, err := dbs.OpenDB("lachesis")
		if err != nil {
			return err
		}
		err = moveDB(skipkeys.Wrap(lachesisDB, MetadataPrefix), newDB, "lachesis", chaindataDir)
		newDB.Close()
		lachesisDB.Close()
		if err != nil {
			return err
		}
		lachesisDB.Drop()

		// move lachesis-%d and gossip-%d DBs
		for _, name := range oldDBs.Names() {
			if strings.HasPrefix(name, "lachesis-") || strings.HasPrefix(name, "gossip-") {
				oldDB, err := oldDBs.OpenDB(name)
				if err != nil {
					return err
				}
				newDB, err = dbs.OpenDB(name)
				if err != nil {
					return err
				}
				err = moveDB(skipkeys.Wrap(oldDB, MetadataPrefix), newDB, name, chaindataDir)
				newDB.Close()
				oldDB.Close()
				if err != nil {
					return err
				}
				oldDB.Drop()
			}
		}

		// move gossip DB
		gossipDB, err := oldDBs.OpenDB("gossip")
		if err != nil {
			return err
		}
		if err = func() error {
			defer gossipDB.Close()

			// move logs
			newDB, err = dbs.OpenDB("evm-logs/r")
			if err != nil {
				return err
			}
			err = moveDB(table.New(gossipDB, []byte("Lr")), newDB, "gossip/Lr", chaindataDir)
			newDB.Close()
			if err != nil {
				return err
			}

			newDB, err = dbs.OpenDB("evm-logs/t")
			err = moveDB(table.New(gossipDB, []byte("Lt")), newDB, "gossip/Lt", chaindataDir)
			newDB.Close()
			if err != nil {
				return err
			}

			// skip 0 prefix, as it contains flushID
			for b := 1; b <= 0xff; b++ {
				if b == int('L') {
					// logs are already moved above
					continue
				}
				if isEmptyDB(table.New(gossipDB, []byte{byte(b)})) {
					continue
				}
				if b == int('M') || b == int('r') || b == int('x') || b == int('X') {
					newDB, err = dbs.OpenDB("evm/" + string([]byte{byte(b)}))
				} else {
					newDB, err = dbs.OpenDB("gossip/" + string([]byte{byte(b)}))
				}
				if err != nil {
					return err
				}
				err = moveDB(table.New(gossipDB, []byte{byte(b)}), newDB, fmt.Sprintf("gossip/%c", rune(b)), chaindataDir)
				newDB.Close()
				if err != nil {
					return err
				}
			}
			return nil
		}(); err != nil {
			return err
		}
		gossipDB.Drop()
	}

	return nil
}
