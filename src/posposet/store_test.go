package posposet

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/dgraph-io/badger"
	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
)

func Test_IntToBytes(t *testing.T) {
	assertar := assert.New(t)

	for _, n1 := range []uint64{
		0,
		9,
		0xFFFFFFFFFFFFFF,
		47528346792,
	} {
		b := intToBytes(n1)
		n2 := bytesToInt(b)
		assertar.Equal(n1, n2)
	}
}

/*
 * bench:
 */

func BenchmarkStoreWithCache(b *testing.B) {
	benchmarkStore(b, true)
}

func BenchmarkNoCachedStore(b *testing.B) {
	benchmarkStore(b, false)
}

func benchmarkStore(b *testing.B, cached bool) {
	dir, err := ioutil.TempDir("", "poset-bench")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			panic(err)
		}
	}()

	opts := badger.DefaultOptions
	opts.Dir = dir
	opts.ValueDir = dir
	opts.SyncWrites = false
	ondisk, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	input := NewEventStore(kvdb.NewBadgerDatabase(ondisk), cached)
	store := NewStore(kvdb.NewBadgerDatabase(ondisk), cached)

	nodes, nodesEvents := GenEventsByNode(5, 100*b.N, 3)
	poset := benchPoset(nodes, input, store, cached)

	b.ResetTimer()

	for _, events := range nodesEvents {
		for _, e := range events {
			input.SetEvent(e.Event)
			poset.PushEventSync(e.Hash())
		}
	}
}

func benchPoset(nodes []hash.Peer, input EventSource, store *Store, cached bool) *Poset {
	balances := make(map[hash.Peer]uint64, len(nodes))
	for _, addr := range nodes {
		balances[addr] = uint64(1)
	}

	if err := store.ApplyGenesis(balances); err != nil {
		panic(err)
	}

	poset := New(store, input)
	poset.Bootstrap()

	return poset
}
