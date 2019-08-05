package posposet

import (
	"fmt"
	"go.etcd.io/bbolt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
)

/*
 * bench:
 */

func BenchmarkStore(b *testing.B) {
	logger.SetTestMode(b)

	benchmarkStore(b)
}

func benchmarkStore(b *testing.B) {
	dir, err := ioutil.TempDir("", "poset-bench")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			panic(err)
		}
	}()

	// open history DB
	pathDb := filepath.Join(dir, "lachesis.bolt")
	db, err := bbolt.Open(pathDb, 0600, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	dbWrapper := kvdb.NewCacheWrapper(kvdb.NewBoltDatabase(db))

	// epoch DB
	prevTemp := 1
	var prevTempDb *bbolt.DB
	var prevTempDbWrapper *kvdb.CacheWrapper
	newTempDb := func() kvdb.Database {
		// close prev temp DB
		if prevTempDb != nil {
			err := prevTempDb.Close()
			if err != nil {
				b.Fatal(err)
			}
			prevTempDb = nil
		}
		// open new one
		pathTemp := filepath.Join(dir, fmt.Sprintf("lachesis.%d.tmp.bolt", prevTemp))
		prevTempDb, err = bbolt.Open(pathTemp, 0600, nil)
		if err != nil {
			panic(err)
		}
		// counter
		prevTemp += 1

		prevTempDbWrapper = kvdb.NewCacheWrapper(kvdb.NewBoltDatabase(prevTempDb))
		return prevTempDbWrapper
	}

	input := NewEventStore(dbWrapper)
	defer input.Close()

	store := NewStore(dbWrapper, newTempDb)
	defer input.Close()

	nodes := inter.GenNodes(5)

	p := benchPoset(nodes, input, store)

	// flushes both epoch DB and history DB
	flushAll := func() {
		err := dbWrapper.Flush()
		if err != nil {
			b.Fatal(err)
		}
		err = prevTempDbWrapper.Flush()
		if err != nil {
			b.Fatal(err)
		}
	}

	// run test with random DAG, N + 1 epochs long
	b.ResetTimer()
	maxEpoch := idx.SuperFrame(b.N) + 1
	for epoch := idx.SuperFrame(1); epoch <= maxEpoch; epoch++ {
		buildEvent := func(e *inter.Event) *inter.Event {
			e.Epoch = epoch
			return p.Prepare(e)
		}
		onNewEvent := func(e *inter.Event) {
			input.SetEvent(e)
			p.PushEventSync(e.Hash())

			if (dbWrapper.NotFlushedSizeEst() + prevTempDbWrapper.NotFlushedSizeEst()) >= 1024*1024 {
				flushAll()
			}
		}

		_ = inter.GenEventsByNode(nodes, int(SuperFrameLen*3), 3, buildEvent, onNewEvent)
	}

	flushAll()
}

func benchPoset(nodes []hash.Peer, input EventSource, store *Store) *Poset {
	balances := make(map[hash.Peer]inter.Stake, len(nodes))
	for _, addr := range nodes {
		balances[addr] = inter.Stake(1)
	}

	if err := store.ApplyGenesis(balances, genesisTestTime); err != nil {
		panic(err)
	}

	poset := New(store, input)
	poset.newBlockCh = nil
	poset.Bootstrap()

	return poset
}
