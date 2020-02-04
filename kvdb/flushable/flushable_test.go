package flushable

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"strconv"
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
	tries := 60            // number of test iterations
	opsPerIter := 0x140    // max number of put/delete ops per iteration
	dictSize := opsPerIter // number of different words

	dir, err := ioutil.TempDir("", "test-flushable")
	if err != nil {
		panic(fmt.Sprintf("can't create temporary directory %s: %v", dir, err))
	}
	disk := leveldb.NewProducer(dir)

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

	for try := 0; try < tries; try++ {

		// random put/delete operations
		putDeleteRandom := func() {
			for j := 0; j < tablesNum; j++ {
				var batches []ethdb.Batch
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

				got := 0
				for ; it.Next(); got++ {
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

func TestFlushableParallel(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-flushable")
	if err != nil {
		panic(fmt.Sprintf("can't create temporary directory %s: %v", dir, err))
	}
	disk := leveldb.NewProducer(dir)

	leveldb2 := disk.OpenDb("2")
	defer leveldb2.Drop()
	defer leveldb2.Close()

	dbLdb := Wrap(leveldb2)
	baseLdb := table.New(dbLdb, []byte{})

	assertar := assert.New(t)

	i := 128
	// Test with i parallel goroutines
	wg := sync.WaitGroup{}
	for j := 0; j < i; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_loopPutGetSameData(assertar, baseLdb, 1000)

			err := dbLdb.Flush()
			assertar.True(err == nil, "Error flush data to DB")
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()

			_loopPutGetDiffData(assertar, baseLdb, 1000)

			err := dbLdb.Flush()
			assertar.True(err == nil, "Error flush data to DB")
		}()
	}
	wg.Wait()
}

func TestFlushableParallelTableLocal(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-flushable")
	if err != nil {
		panic(fmt.Sprintf("can't create temporary directory %s: %v", dir, err))
	}
	disk := leveldb.NewProducer(dir)

	leveldb2 := disk.OpenDb("2")
	defer leveldb2.Drop()
	defer leveldb2.Close()

	dbLdb := Wrap(leveldb2)

	assertar := assert.New(t)

	i := 128
	// Test with i parallel goroutines
	wg := sync.WaitGroup{}
	for j := 0; j < i; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			baseLdb := table.New(dbLdb, []byte{})
			_loopPutGetSameData(assertar, baseLdb, 1000)

			err := dbLdb.Flush()
			assertar.True(err == nil, "Error flush data to DB")
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()

			baseLdb := table.New(dbLdb, []byte{})
			_loopPutGetDiffData(assertar, baseLdb, 1000)

			err := dbLdb.Flush()
			assertar.True(err == nil, "Error flush data to DB")
		}()
	}
	wg.Wait()
}

func TestFlushableIteratorParallel(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-flushable")
	if err != nil {
		panic(fmt.Sprintf("can't create temporary directory %s: %v", dir, err))
	}
	disk := leveldb.NewProducer(dir)

	leveldb2 := disk.OpenDb("2")
	defer leveldb2.Drop()
	defer leveldb2.Close()

	dbLdb := Wrap(leveldb2)
	baseLdb := table.New(dbLdb, []byte{})

	assertar := assert.New(t)


	// Prepare data
	_loopPutGetDiffData(assertar, baseLdb, 1000)
	dbLdb.Flush()

	i := 128
	// Test with i parallel goroutines
	wg := sync.WaitGroup{}
	for j := 0; j < i; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			it := dbLdb.NewIterator()
			defer it.Release()

			expectPairs := map[string][]byte{}

			got := 0
			for ; it.Next(); got++ {
				expectPairs[string(it.Key())] = it.Value()
				assertar.False(t.Failed(), "Parallel iterator failed")
			}

			assertar.NoError(it.Error())

			assertar.Equal(len(expectPairs), got) // check that we've got the same num of pairs
		}()
	}
	wg.Wait()
}

func BenchmarkFlushable_PutGet(b *testing.B) {
	dir, err := ioutil.TempDir("", "test-flushable")
	if err != nil {
		panic(fmt.Sprintf("can't create temporary directory %s: %v", dir, err))
	}
	disk := leveldb.NewProducer(dir)

	// open raw databases
	leveldb2 := disk.OpenDb("2")
	defer leveldb2.Drop()
	defer leveldb2.Close()

	dbLdb := Wrap(leveldb2)
	baseLdb := table.New(dbLdb, []byte{})

	allThreads := 16384
	for i := 1; i <= allThreads; i*=2 {
		pNum := i
		b.Run("Sequenced "+strconv.FormatInt(int64(allThreads/pNum), 10)+" parallel "+strconv.FormatInt(int64(pNum), 10), func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_parallelBenchmarkPutGet(baseLdb, pNum, allThreads)
				dbLdb.Flush()
			}
		})
	}
}

var flushCounter int

func BenchmarkFlushable_PutGet_WithFlush(b *testing.B) {
	dir, err := ioutil.TempDir("", "test-flushable")
	if err != nil {
		panic(fmt.Sprintf("can't create temporary directory %s: %v", dir, err))
	}
	disk := leveldb.NewProducer(dir)

	// open raw databases
	leveldb2 := disk.OpenDb("2")
	defer leveldb2.Drop()
	defer leveldb2.Close()

	dbLdb := Wrap(leveldb2)
	baseLdb := table.New(dbLdb, []byte{})

	for flushAfter := 1; flushAfter <= 1000; flushAfter *= 10 {
		flushCounter = flushAfter
		for allThreads := 16384; allThreads > 1024; allThreads/=2 {
			for i := 1; i <= allThreads; i *= 2 {
				pNum := i
				b.Run("Flush every "+strconv.FormatInt(int64(flushAfter), 10)+
					" sequenced "+strconv.FormatInt(int64(allThreads/pNum), 10)+
					" parallel "+strconv.FormatInt(int64(pNum), 10), func(b *testing.B) {
					for n := 0; n < b.N; n++ {
						_parallelBenchmarkPutGetFlush(baseLdb, dbLdb, pNum, allThreads, flushAfter)
					}
				})
			}
		}
	}
}

func _parallelBenchmarkPutGet(tbl *table.Table, pNum, allThreads int) {
	wg := sync.WaitGroup{}

	r := rand.New(rand.NewSource(0))
	testKey := big.NewInt(r.Int63()).Bytes()
	testVal := big.NewInt(r.Int63()).Bytes()

	seqNum := allThreads / pNum

	for i := 0; i < pNum; i++ {
		wg.Add(1)
		go func(){
			defer wg.Done()

			for j := 0; j < seqNum; j++ {
				_ = tbl.Put(testKey, testVal)
				_, _ = tbl.Get(testKey)
			}
		}()
	}
	wg.Wait()
}

func _parallelBenchmarkPutGetFlush(tbl *table.Table, db *Flushable, pNum, allThreads, flushAfter int) {
	wg := sync.WaitGroup{}

	r := rand.New(rand.NewSource(1))
	mu := sync.Mutex{}

	seqNum := allThreads / pNum

	for i := 0; i < pNum; i++ {
		wg.Add(1)
		go func(){
			defer wg.Done()

			for j := 0; j < seqNum; j++ {
				mu.Lock()
				testKey := big.NewInt(r.Int63()).Bytes()
				testVal := big.NewInt(r.Int63()).Bytes()
				mu.Unlock()

				_ = tbl.Put(testKey, testVal)
				_, _ = tbl.Get(testKey)

				flushCounter--
				if flushCounter <= 0 {
					db.Flush()
					flushCounter = flushAfter
				}
			}
		}()
	}
	wg.Wait()
}

func _loopPutGetSameData(assertar *assert.Assertions, tbl *table.Table, loopCount int) {
	r := rand.New(rand.NewSource(0))

	testKey := big.NewInt(r.Int63()).Bytes()
	testVal := big.NewInt(r.Int63()).Bytes()

	for i := 0; i < loopCount; i++ {
		err := tbl.Put(testKey, testVal)
		assertar.True(err == nil, "Error put data to DB")
		_, err = tbl.Get(testKey)
		assertar.True(err == nil, "Error get data from DB")
	}
}

func _loopPutGetDiffData(assertar *assert.Assertions, tbl *table.Table, loopCount int) {
	r := rand.New(rand.NewSource(0))

	for i := 0; i < loopCount; i++ {
		testKey := big.NewInt(r.Int63()).Bytes()
		testVal := big.NewInt(r.Int63()).Bytes()

		err := tbl.Put(testKey, testVal)
		assertar.True(err == nil, "Error put data to DB")
		_, err = tbl.Get(testKey)
		assertar.True(err == nil, "Error get data from DB")
	}
}
