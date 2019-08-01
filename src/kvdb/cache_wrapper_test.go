package kvdb

import (
	"github.com/Fantom-foundation/go-lachesis/src/common"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestForEach2(t *testing.T) {
	tries := 15            // number of test iterations
	opsPerIter := 0x200    // max number of put/delete ops per iteration
	dictSize := opsPerIter // number of different words

	// open raw databases
	bbolt1, freeBbolt1 := bboltDB("1")
	defer freeBbolt1()

	bbolt2, freeBbolt2 := bboltDB("2")
	defer freeBbolt2()

	// create wrappers
	dbs := map[string]Database{
		"bbolt":  NewBoltDatabase(bbolt1),
		"memory": NewMemDatabase(),
	}

	flushableDbs := map[string]FlushableDatabase{
		"cache-over-bbolt":  NewCacheWrapper(NewBoltDatabase(bbolt2)),
		"cache-over-memory": NewCacheWrapper(NewMemDatabase()),
	}

	dbsTables := [][]Database{
		{
			dbs["bbolt"],
			dbs["bbolt"].NewTable([]byte{0, 1}),
			dbs["bbolt"].NewTable([]byte{0}),
			dbs["bbolt"].NewTable([]byte{0}).NewTable(common.Hex2Bytes("fffffffffffffffffffffffffffffffffffffe")),
			dbs["bbolt"].NewTable([]byte{0}).NewTable(common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffff")),
		},
		{
			dbs["memory"],
			dbs["memory"].NewTable([]byte{0, 1}),
			dbs["memory"].NewTable([]byte{0}),
			dbs["memory"].NewTable([]byte{0}).NewTable(common.Hex2Bytes("fffffffffffffffffffffffffffffffffffffe")),
			dbs["memory"].NewTable([]byte{0}).NewTable(common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffff")),
		},
	}

	flushableDbsTables := [][]FlushableDatabase{
		{
			flushableDbs["cache-over-bbolt"],
			flushableDbs["cache-over-bbolt"].NewTableFlushable([]byte{0, 1}),
			flushableDbs["cache-over-bbolt"].NewTableFlushable([]byte{0}),
			flushableDbs["cache-over-bbolt"].NewTableFlushable([]byte{0}).NewTableFlushable(common.Hex2Bytes("fffffffffffffffffffffffffffffffffffffe")),
			flushableDbs["cache-over-bbolt"].NewTableFlushable([]byte{0}).NewTableFlushable(common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffff")),
		},
		{
			flushableDbs["cache-over-memory"],
			flushableDbs["cache-over-memory"].NewTableFlushable([]byte{0, 1}),
			flushableDbs["cache-over-memory"].NewTableFlushable([]byte{0}),
			flushableDbs["cache-over-memory"].NewTableFlushable([]byte{0}).NewTableFlushable(common.Hex2Bytes("fffffffffffffffffffffffffffffffffffffe")),
			flushableDbs["cache-over-memory"].NewTableFlushable([]byte{0}).NewTableFlushable(common.Hex2Bytes("ffffffffffffffffffffffffffffffffffffff")),
		},
	}

	assertar := assert.New(t)
	assertar.Equal(len(dbsTables), len(flushableDbsTables))
	assertar.Equal(len(dbsTables[0]), len(flushableDbsTables[0]))

	groupsNum := len(dbsTables)
	tablesNum := len(dbsTables[0])

	// use the same seed for determinism
	r := rand.New(rand.NewSource(0))

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
		b := append(prefixes[i%len(prefixes)], big.NewInt(r.Int63()).Bytes()...)
		dict = append(dict, b)
	}

	for try := 0; try < tries; try += 1 {
		// random pet/delete operations
		putDeleteRandom := func() {
			for j := 0; j < tablesNum; j++ {
				var batches []Batch
				for i := 0; i < groupsNum; i++ {
					batches = append(batches, dbsTables[i][j].NewBatch())
					batches = append(batches, flushableDbsTables[i][j].NewBatch())
				}

				ops := 1 + r.Intn(opsPerIter)
				for p := 0; p < ops; p++ {
					var pair kv
					if r.Intn(2) == 0 { // put
						pair = kv{
							k: dict[r.Intn(len(dict))],
							v: dict[r.Intn(len(dict))],
						}
					} else { // delete
						pair = kv{
							k: dict[r.Intn(len(dict))],
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
			prefix = []byte{0, 0, 0, 0, 0, 0}
		}
		stopAfter := 1 + r.Intn(len(dict))

		for j := 0; j < tablesNum; j++ {
			expectPairs := []kv{}
			testForEach := func(db Database, first bool) {
				got := 0
				assertar.NoError(db.ForEach(prefix, func(key, val []byte) bool {
					assertar.NotEqual(got, stopAfter) // check that return true/false works here

					if first {
						expectPairs = append(expectPairs, kv{
							k: key,
							v: val,
						})
					} else {
						assertar.NotEqual(got, len(expectPairs)) // check that we've for the same num of values
						assertar.Equal(key, expectPairs[got].k)
						assertar.Equal(val, expectPairs[got].v)
					}

					got += 1
					return got < stopAfter
				}))

				assertar.Equal(got, len(expectPairs)) // check that we've got the same num of pairs
			}

			// check that all groups return the same result
			for i := 0; i < groupsNum; i++ {
				testForEach(dbsTables[i][j], i == 0)
				testForEach(flushableDbsTables[i][j], false)
			}
		}

		// try to get random values
		ops := r.Intn(opsPerIter)
		for p := 0; p < ops; p++ {
			key := dict[r.Intn(len(dict))]

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
