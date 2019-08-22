package table

import (
	"bytes"
	"fmt"
	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/flushable"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/leveldb"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb/memorydb"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/dgraph-io/badger"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
)

func tempLevelDB(name string) *leveldb.Database {
	dir, err := ioutil.TempDir("", "flushable-test"+name)
	if err != nil {
		panic(fmt.Sprintf("can't create temporary directory: %v", err))
	}

	drop := func() error {
		_ = os.RemoveAll(dir)
		return nil
	}

	diskdb, err := leveldb.New(dir, 16, 0, "", nil, drop)
	if err != nil {
		panic(fmt.Sprintf("can't create temporary database: %v", err))
	}
	return diskdb
}

func TestTable(t *testing.T) {
	prefix0 := map[string][]byte{
		"00": []byte{0},
		"01": []byte{0, 1},
		"02": []byte{0, 1, 2},
		"03": []byte{0, 1, 2, 3},
	}
	prefix1 := map[string][]byte{
		"10": []byte{0, 1, 2, 3, 4},
	}
	testData := join(prefix0, prefix1)

	// open raw databases
	leveldb1 := tempLevelDB("1")
	defer leveldb1.Drop()
	defer leveldb1.Close()

	leveldb2 := tempLevelDB("2")
	defer leveldb2.Drop()
	defer leveldb2.Close()

	for name, db := range map[string]kvdb.KeyValueStore{
		"memory":                       memorydb.New(),
		"leveldb":                      leveldb1,
		"cache-over-leveldb":           flushable.New(leveldb2),
		"cache-over-cache-over-memory": flushable.New(memorydb.New()),
	} {
		t.Run(name, func(t *testing.T) {
			assertar := assert.New(t)

			// tables
			t1 := New(db, []byte("t1"))
			tables := map[string]kvdb.KeyValueStore{
				"/t1":   t1,
				"/t1/x": t1.NewTable([]byte("x")),
				"/t2":   New(db, []byte("t2")),
			}

			// write
			for name, t := range tables {
				for k, v := range testData {
					err := t.Put([]byte(k), v)
					if !assertar.NoError(err, name) {
						return
					}
				}
			}

			// read
			for name, t := range tables {

				for pref, count := range map[string]int{
					"0": len(prefix0),
					"1": len(prefix1),
					"":  len(prefix0) + len(prefix1),
				} {
					got := 0
					var prevKey []byte

					it := t.NewIteratorWithPrefix([]byte(pref))
					for it.Next() {
						if prevKey == nil {
							prevKey = common.CopyBytes(it.Key())
						} else {
							assertar.Equal(1, bytes.Compare(it.Key(), prevKey))
						}
						got++
						assertar.Equal(
							testData[string(it.Key())],
							it.Value(),
							name+": "+string(it.Key()),
						)
					}

					if !assertar.NoError(it.Error()) {
						return
					}

					it.Release()

					if !assertar.Equal(count, got) {
						return
					}
				}
			}
		})
	}
}

func join(aa ...map[string][]byte) map[string][]byte {
	res := make(map[string][]byte)
	for _, a := range aa {
		for k, v := range a {
			res[k] = v
		}
	}

	return res
}

func bboltDB(dir string) (db *bbolt.DB, drop func()) {
	dir, err := ioutil.TempDir("", "kvdb"+dir)
	if err != nil {
		panic(err)
	}
	f := filepath.Join(dir, "bbolt.db")

	db, err = bbolt.Open(f, 0600, nil)
	if err != nil {
		panic(err)
	}

	drop = func() {
		_ = db.Close()
		_ = os.RemoveAll(dir)
	}

	return
}

func badgerDB(dir string) (db *badger.DB, drop func()) {
	dir, err := ioutil.TempDir("", "kvdb"+dir)
	if err != nil {
		panic(err)
	}

	opts := badger.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir

	db, err = badger.Open(opts)
	if err != nil {
		panic(err)
	}

	drop = func() {
		_ = db.Close()
		_ = os.RemoveAll(dir)
	}

	return
}
