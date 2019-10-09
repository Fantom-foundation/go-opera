package table

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/flushable"
	"github.com/Fantom-foundation/go-lachesis/kvdb/leveldb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
)

func tempLevelDB(name string) *leveldb.Database {
	dir, err := ioutil.TempDir("", "flushable-test"+name)
	if err != nil {
		panic(fmt.Sprintf("can't create temporary directory: %v", err))
	}

	drop := func() {
		err := os.RemoveAll(dir)
		if err != nil {
			panic(err)
		}
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
		"cache-over-leveldb":           flushable.Wrap(leveldb2),
		"cache-over-cache-over-memory": flushable.Wrap(memorydb.New()),
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
					defer it.Release()
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
