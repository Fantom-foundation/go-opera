package flushable

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/kvdb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/leveldb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/memorydb"
	"github.com/Fantom-foundation/go-lachesis/kvdb/table"
)

func TestFlushable(t *testing.T) {
	assertar := assert.New(t)

	tries := 60            // number of test iterations
	opsPerIter := 0x140    // max number of put/delete ops per iteration
	dictSize := opsPerIter // number of different words

	disk := dbProducer("TestFlushable")

	// open raw databases
	leveldb1 := disk.OpenDb("1")
	defer leveldb1.Drop()
	defer leveldb1.Close()

	leveldb2 := disk.OpenDb("2")
	defer leveldb2.Drop()
	defer leveldb2.Close()

	// create wrappers
	dbs := map[string]kvdb.KeyValueStore{
		"leveldb": leveldb1,
		"memory":  memorydb.New(),
	}

	flushableDbs := map[string]*Flushable{
		"cache-over-leveldb": Wrap(leveldb2),
		"cache-over-memory":  Wrap(memorydb.New()),
	}

	baseLdb := table.New(dbs["leveldb"], []byte{})
	baseMem := table.New(dbs["memory"], []byte{})

	dbsTables := [][]ethdb.KeyValueStore{
		{
			dbs["leveldb"],
			baseLdb.NewTable([]byte{0, 1}),
			baseLdb.NewTable([]byte{0}).NewTable(common.Hex2Bytes("ffffffffffffffffffffffffffffffffffff")),
		},
		{
			dbs["memory"],
			baseMem.NewTable([]byte{0, 1}),
			baseMem.NewTable([]byte{0}).NewTable(common.Hex2Bytes("ffffffffffffffffffffffffffffffffffff")),
		},
	}

	baseLdb = table.New(flushableDbs["cache-over-leveldb"], []byte{})
	baseMem = table.New(flushableDbs["cache-over-memory"], []byte{})
	flushableDbsTables := [][]kvdb.KeyValueStore{
		{
			flushableDbs["cache-over-leveldb"],
			baseLdb.NewTable([]byte{0, 1}),
			baseLdb.NewTable([]byte{0}).NewTable(common.Hex2Bytes("ffffffffffffffffffffffffffffffffffff")),
		},
		{
			flushableDbs["cache-over-memory"],
			baseMem.NewTable([]byte{0, 1}),
			baseMem.NewTable([]byte{0}).NewTable(common.Hex2Bytes("ffffffffffffffffffffffffffffffffffff")),
		},
	}

	assertar.Equal(len(dbsTables), len(flushableDbsTables))
	assertar.Equal(len(dbsTables[0]), len(flushableDbsTables[0]))

	groupsNum := len(dbsTables)
	tablesNum := len(dbsTables[0])

	// use the same seed for determinism
	rand := rand.New(rand.NewSource(0))

	// words dictionary
	prefixes := [][]byte{
		{},
		{0},
		{0x1},
		{0x22},
		{0x33},
		{0x11},
		{0x11, 0x22},
		{0x11, 0x23},
		{0x11, 0x22, 0x33},
		{0x11, 0x22, 0x34},
	}
	dict := [][]byte{}
	for i := 0; i < dictSize; i++ {
		b := append(prefixes[i%len(prefixes)], big.NewInt(rand.Int63()).Bytes()...)
		dict = append(dict, b)
	}

	for try := 0; try < tries; try++ {

		// random put/delete operations
		putDeleteRandom := func() {
			for j := 0; j < tablesNum; j++ {
				var batches []ethdb.Batch
				for i := 0; i < groupsNum; i++ {
					batches = append(batches, dbsTables[i][j].NewBatch())
					batches = append(batches, flushableDbsTables[i][j].NewBatch())
				}

				ops := 1 + rand.Intn(opsPerIter)
				for p := 0; p < ops; p++ {
					var pair kv
					if rand.Intn(2) == 0 { // put
						pair = kv{
							k: dict[rand.Intn(len(dict))],
							v: dict[rand.Intn(len(dict))],
						}
					} else { // delete
						pair = kv{
							k: dict[rand.Intn(len(dict))],
							v: nil,
						}
					}

					for _, batch := range batches {
						if pair.v != nil {
							assertar.NoError(batch.Put(pair.k, pair.v))
						} else {
							assertar.NoError(batch.Delete(pair.k))
						}
					}
				}

				for _, batch := range batches {
					size := batch.ValueSize()
					assertar.NotEqual(0, size)
					assertar.NoError(batch.Write())
					assertar.Equal(size, batch.ValueSize())
					batch.Reset()
					assertar.Equal(0, batch.ValueSize())
				}
			}
		}
		// put/delete values
		putDeleteRandom()

		// flush
		for _, db := range flushableDbs {
			if try == 0 && !assertar.NotEqual(0, db.NotFlushedPairs()) {
				return
			}
			assertar.NoError(db.Flush())
			assertar.Equal(0, db.NotFlushedPairs())
		}

		// put/delete values (not flushed)
		putDeleteRandom()

		// try to ForEach random prefix
		prefix := prefixes[try%len(prefixes)]
		if try == 1 {
			prefix = []byte{0, 0, 0, 0, 0, 0} // not existing prefix
		}

		for j := 0; j < tablesNum; j++ {
			expectPairs := []kv{}

			testForEach := func(db ethdb.KeyValueStore, first bool) {

				var it ethdb.Iterator
				if try%3 == 0 {
					it = db.NewIterator()
				} else if try%3 == 1 {
					it = db.NewIteratorWithPrefix(prefix)
				} else {
					it = db.NewIteratorWithStart(prefix)
				}
				defer it.Release()

				var got int

				for got = 0; it.Next(); got++ {
					if first {
						expectPairs = append(expectPairs, kv{
							k: common.CopyBytes(it.Key()),
							v: common.CopyBytes(it.Value()),
						})
					} else {
						assertar.NotEqual(len(expectPairs), got, try) // check that we've for the same num of values
						if t.Failed() {
							return
						}
						assertar.Equal(expectPairs[got].k, it.Key(), try)
						assertar.Equal(expectPairs[got].v, it.Value(), try)
					}
				}

				if !assertar.NoError(it.Error()) {
					return
				}

				assertar.Equal(len(expectPairs), got) // check that we've got the same num of pairs
			}

			// check that all groups return the same result
			for i := 0; i < groupsNum; i++ {
				testForEach(dbsTables[i][j], i == 0)
				if t.Failed() {
					return
				}
				testForEach(flushableDbsTables[i][j], false)
				if t.Failed() {
					return
				}
			}
		}

		// try to get random values
		ops := rand.Intn(opsPerIter)
		for p := 0; p < ops; p++ {
			key := dict[rand.Intn(len(dict))]

			for j := 0; j < tablesNum; j++ {
				// get values for first group, so we could check that all groups return the same result
				ok, _ := dbsTables[0][j].Has(key)
				vl, _ := dbsTables[0][j].Get(key)

				// check that all groups return the same result
				for i := 0; i < groupsNum; i++ {
					ok1, err := dbsTables[i][j].Has(key)
					assertar.NoError(err)
					vl1, err := dbsTables[i][j].Get(key)
					assertar.NoError(err)

					ok2, err := flushableDbsTables[i][j].Has(key)
					assertar.NoError(err)
					vl2, err := flushableDbsTables[i][j].Get(key)
					assertar.NoError(err)

					assertar.Equal(ok1, ok2)
					assertar.Equal(vl1, vl2)
					assertar.Equal(ok1, ok)
					assertar.Equal(vl1, vl)
				}
			}
		}

		if t.Failed() {
			return
		}
	}
}

func TestFlushableIterator(t *testing.T) {
	assertar := assert.New(t)

	disk := dbProducer("TestFlushableIterator")

	leveldb := disk.OpenDb("1")
	defer leveldb.Drop()
	defer leveldb.Close()

	flushable1 := Wrap(leveldb)
	flushable2 := Wrap(leveldb)

	allkeys := [][]byte{
		{0x11, 0x00},
		{0x12, 0x00},
		{0x13, 0x00},
		{0x14, 0x00},
		{0x15, 0x00},
		{0x16, 0x00},
		{0x17, 0x00},
		{0x18, 0x00},
		{0x19, 0x00},
		{0x1a, 0x00},
		{0x1b, 0x00},
		{0x1c, 0x00},
		{0x1d, 0x00},
		{0x1e, 0x00},
		{0x1f, 0x00},
	}

	veryFirstKey := allkeys[0]
	veryLastKey := allkeys[len(allkeys)-1]
	expected := allkeys[1 : len(allkeys)-1]

	for _, key := range expected {
		leveldb.Put(key, []byte("in-order"))
	}

	flushable2.Put(veryFirstKey, []byte("first"))
	flushable2.Put(veryLastKey, []byte("last"))

	it := flushable1.NewIterator()
	defer it.Release()

	err := flushable2.Flush()
	if !assertar.NoError(err) {
		return
	}

	for i := 0; it.Next(); i++ {
		if !assertar.Equal(expected[i], it.Key()) ||
			!assertar.Equal([]byte("in-order"), it.Value()) {
			break
		}
	}
}

func BenchmarkFlushable(b *testing.B) {
	disk := dbProducer("BenchmarkFlushable")

	leveldb := disk.OpenDb("1")
	defer leveldb.Drop()
	defer leveldb.Close()

	flushable := Wrap(leveldb)

	const recs = 10000

	for _, flushPeriod := range []int{0, 1, 10, 100, 1000} {
		for goroutines := 1; goroutines <= recs/2; goroutines *= 2 {
			name := fmt.Sprintf(
				"%d goroutines with flush every %d ops",
				goroutines, flushPeriod)
			b.Run(name, func(b *testing.B) {
				benchmarkFlushable(flushable, goroutines, recs*b.N, flushPeriod)
			})
		}
	}
}

func benchmarkFlushable(db *Flushable, goroutines, recs, flushPeriod int) {
	var ops = recs / goroutines

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()

			rand := rand.New(rand.NewSource(int64(i)))
			flushOffset := flushPeriod * i / goroutines

			for op := 0; op < ops; op++ {
				key := big.NewInt(rand.Int63()).Bytes()
				val := big.NewInt(rand.Int63()).Bytes()

				err := db.Put(key, val)
				if err != nil {
					panic(err)
				}
				got, err := db.Get(key)
				if err != nil {
					panic(err)
				}
				if !bytes.Equal(val, got) {
					panic("invalid value")
				}

				if flushPeriod != 0 && (op+flushOffset)%flushPeriod == 0 {
					db.Flush()
				}
			}
		}(i)
	}
	wg.Wait()
}

func dbProducer(name string) kvdb.DbProducer {
	dir, err := ioutil.TempDir("", name)
	if err != nil {
		panic(err)
	}
	return leveldb.NewProducer(dir)
}
