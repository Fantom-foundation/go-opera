package kvdb

import (
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

	bbolt, freeBbolt := bboltDB()
	defer freeBbolt()

	badger, freeBadger := badgerDB()
	defer freeBadger()

	for name, db := range map[string]Database{
		"memory": NewMemDatabase(),
		"bbolt":  NewBoltDatabase(bbolt),
		"badger": NewBadgerDatabase(badger),
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
					err := t.ForEach([]byte(pref), func(key, val []byte) bool {
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

func bboltDB() (db *bbolt.DB, free func()) {
	dir, err := ioutil.TempDir("", "kvdb")
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

func badgerDB() (db *badger.DB, free func()) {
	dir, err := ioutil.TempDir("", "kvdb")
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
