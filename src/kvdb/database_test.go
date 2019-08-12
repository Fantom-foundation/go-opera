package kvdb

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/dgraph-io/badger"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
)

func TestForEach(t *testing.T) {
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

	bbolt1, freeBbolt1 := bboltDB("1")
	defer freeBbolt1()

	bbolt2, freeBbolt2 := bboltDB("2")
	defer freeBbolt2()

	badger1, freeBadger1 := badgerDB("1")
	defer freeBadger1()

	for name, db := range map[string]Database{
		"memory":                       NewMemDatabase(),
		"bbolt":                        NewBoltDatabase(bbolt1),
		"badger":                       NewBadgerDatabase(badger1),
		"cache-over-bbolt":             NewCacheWrapper(NewBoltDatabase(bbolt2)),
		"cache-over-cache-over-memory": NewCacheWrapper(NewCacheWrapper(NewMemDatabase())),
	} {
		t.Run(name, func(t *testing.T) {
			assertar := assert.New(t)

			// tables
			t1 := db.NewTable([]byte("t1"))
			tables := map[string]Database{
				"/t1":   t1,
				"/t1/x": t1.NewTable([]byte("x")),
				"/t2":   db.NewTable([]byte("t2")),
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
					err := t.ForEach([]byte(pref), func(key, val []byte) bool {
						if prevKey == nil {
							prevKey = key
						} else {
							assertar.Equal(1, bytes.Compare(key, prevKey))
						}
						got++
						return assertar.Equal(
							testData[string(key)],
							val,
							name+": "+string(key),
						)
					})

					if !assertar.NoError(err) {
						return
					}

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

func bboltDB(dir string) (db *bbolt.DB, free func()) {
	dir, err := ioutil.TempDir("", "kvdb" + dir)
	if err != nil {
		panic(err)
	}
	f := filepath.Join(dir, "bbolt.db")

	db, err = bbolt.Open(f, 0600, nil)
	if err != nil {
		panic(err)
	}

	free = func() {
		_ = os.RemoveAll(dir)
	}

	return
}

func badgerDB(dir string) (db *badger.DB, free func()) {
	dir, err := ioutil.TempDir("", "kvdb" + dir)
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

	free = func() {
		_ = os.RemoveAll(dir)
	}

	return
}
