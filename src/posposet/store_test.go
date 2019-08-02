package posposet

import (
	"github.com/Fantom-foundation/go-lachesis/src/kvdb"
	"go.etcd.io/bbolt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/Fantom-foundation/go-lachesis/src/common"
	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter"
	"github.com/Fantom-foundation/go-lachesis/src/logger"
	"github.com/stretchr/testify/assert"
)

func Test_IntToBytes(t *testing.T) {
	assertar := assert.New(t)

	for _, n1 := range []uint64{
		0,
		9,
		0xFFFFFFFFFFFFFF,
		47528346792,
	} {
		b := common.IntToBytes(n1)
		n2 := common.BytesToInt(b)
		assertar.Equal(n1, n2)
	}
}

/*
 * bench:
 */

func BenchmarkStoreWithCache(b *testing.B) {
	logger.SetTestMode(b)

	benchmarkStore(b, true)
}

func BenchmarkNoCachedStore(b *testing.B) {
	logger.SetTestMode(b)

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

	f := filepath.Join(dir, "lachesis.bolt")
	ondisk, err := bbolt.Open(f, 0600, nil)
	if err != nil {
		panic(err)
	}
	defer ondisk.Close()

	input := NewEventStore(kvdb.NewBoltDatabase(ondisk))
	defer input.Close()
	store := NewStore(kvdb.NewBoltDatabase(ondisk))
	defer input.Close()

	nodes := inter.GenNodes(5)

	p := benchPoset(nodes, input, store, cached)

	buildEvent := func(e *inter.Event) *inter.Event {
		return p.Prepare(e)
	}
	onNewEvent := func(e *inter.Event) {
		input.SetEvent(e)
		p.PushEventSync(e.Hash())
	}

	b.ResetTimer()
	_ = inter.GenEventsByNode(nodes, 100*b.N, 3, buildEvent, onNewEvent)
}

func benchPoset(nodes []hash.Peer, input EventSource, store *Store, cached bool) *Poset {
	balances := make(map[hash.Peer]inter.Stake, len(nodes))
	for _, addr := range nodes {
		balances[addr] = inter.Stake(1)
	}

	if err := store.ApplyGenesis(balances); err != nil {
		panic(err)
	}

	poset := New(store, input)
	poset.Bootstrap()

	return poset
}
