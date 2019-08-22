package poset

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"go.etcd.io/bbolt"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"github.com/Fantom-foundation/go-lachesis/src/lachesis"
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

	var (
		epochCache *kvdb.CacheWrapper
	)
	newDb := func(name string) kvdb.Database {
		path := filepath.Join(dir, fmt.Sprintf("lachesis.%s.bolt", name))
		db, err := bbolt.Open(path, 0600, nil)
		if err != nil {
			panic(err)
		}

		cache := kvdb.NewCacheWrapper(
			kvdb.NewBoltDatabase(
				db,
				nil,
				func() error {
					return os.Remove(path)
				}))
		if name == "epoch" {
			epochCache = cache
		}
		return cache
	}

	// open history DB
	historyCache := kvdb.NewCacheWrapper(newDb("main"))

	input := NewEventStore(historyCache)
	defer input.Close()

	store := NewStore(historyCache, newDb)
	defer store.Close()

	nodes := inter.GenNodes(5)

	p := benchPoset(nodes, input, store)

	// flushes both epoch DB and history DB
	flushAll := func() {
		err := historyCache.Flush()
		if err != nil {
			b.Fatal(err)
		}
		err = epochCache.Flush()
		if err != nil {
			b.Fatal(err)
		}
	}

	p.applyBlock = func(block *inter.Block, stateHash hash.Hash, members pos.Members) (hash.Hash, pos.Members) {
		if block.Index == 1 {
			// move stake from node0 to node1
			members.Set(nodes[0], 0)
			members.Set(nodes[1], 2)
		}
		return stateHash, members
	}

	// run test with random DAG, N + 1 epochs long
	b.ResetTimer()
	maxEpoch := idx.SuperFrame(b.N) + 1
	for epoch := idx.SuperFrame(1); epoch <= maxEpoch; epoch++ {
		r := rand.New(rand.NewSource(int64((epoch))))
		_ = inter.ForEachRandEvent(nodes, int(SuperFrameLen*3), 3, r, inter.ForEachEvent{
			Process: func(e *inter.Event, name string) {
				input.SetEvent(e)
				_ = p.ProcessEvent(e)

				if (historyCache.NotFlushedSizeEst() + epochCache.NotFlushedSizeEst()) >= 1024*1024 {
					flushAll()
				}
			},
			Build: func(e *inter.Event, name string) *inter.Event {
				e.Epoch = epoch
				return p.Prepare(e)
			},
		})
	}

	flushAll()
}

func benchPoset(nodes []hash.Peer, input EventSource, store *Store) *Poset {
	balances := make(map[hash.Peer]pos.Stake, len(nodes))
	for _, addr := range nodes {
		balances[addr] = pos.Stake(1)
	}

	err := store.ApplyGenesis(&lachesis.Genesis{
		Balances: balances,
		Time:     genesisTestTime,
	})
	if err != nil {
		panic(err)
	}

	poset := New(store, input)
	poset.Bootstrap(nil)

	return poset
}
