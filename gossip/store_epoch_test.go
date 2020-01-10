package gossip

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/logger"
)

func TestStoreGetEventHeader(t *testing.T) {
	assertar := assert.New(t)
	logger.SetTestMode(t)

	store := cachedStore()
	expect := fakeEvent()
	h := expect.Hash()

	store.SetEventHeader(expect.Epoch, h, &expect.EventHeaderData)
	got := store.GetEventHeader(expect.Epoch, h)

	assertar.EqualValues(&expect.EventHeaderData, got)
}

func TestStoreEpochStore(t *testing.T) {
	logger.SetTestMode(t)

	dir, err := ioutil.TempDir("", "epochstore-test")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	for name, store := range map[string]*Store{
		"memory": cachedStore(),
		"ondisk": realStore(dir),
	} {
		t.Run(name, func(t *testing.T) {
			testStoreEpochStore(t, store)
		})
		store.Close()
	}
}

func testStoreEpochStore(t *testing.T, store *Store) {
	assertar := assert.New(t)
	var (
		epoch = idx.Epoch(1)
		key   = idx.StakerID(9).Bytes()
		val   = hash.FakeEvent().Bytes()
	)

	es := store.getEpochStore(epoch)
	if !assertar.NotNil(es) {
		return
	}

	err := es.Tips.Put(key, val)
	if !assertar.NoError(err) {
		return
	}

	got, err := es.Tips.Get(key)
	if !assertar.NoError(err) {
		return
	}
	if !assertar.Equal(val, got) {
		return
	}

	store.delEpochStore(epoch)

	got, err = es.Tips.Get(key)
	if !assertar.NoError(err) {
		return
	}
	if !assertar.Nil(got) {
		return
	}

	err = es.Tips.Put(key, val)
	if !assertar.NoError(err) {
		return
	}
}

func BenchmarkStoreGetEventHeader(b *testing.B) {
	logger.SetTestMode(b)

	b.Run("cache on", func(b *testing.B) {
		benchStoreGetEventHeader(b, cachedStore())
	})
	b.Run("cache off", func(b *testing.B) {
		benchStoreGetEventHeader(b, nonCachedStore())
	})
}

func benchStoreGetEventHeader(b *testing.B, store *Store) {
	e := &inter.Event{}
	h := e.Hash()

	store.SetEventHeader(e.Epoch, h, &e.EventHeaderData)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if store.GetEventHeader(e.Epoch, h) == nil {
			b.Fatal("invalid result")
		}
	}
}

func BenchmarkStoreSetEventHeader(b *testing.B) {
	logger.SetTestMode(b)

	b.Run("cache on", func(b *testing.B) {
		benchStoreSetEventHeader(b, cachedStore())
	})
	b.Run("cache off", func(b *testing.B) {
		benchStoreSetEventHeader(b, nonCachedStore())
	})
}

func benchStoreSetEventHeader(b *testing.B, store *Store) {
	e := fakeEvent()
	h := e.Hash()

	for i := 0; i < b.N; i++ {
		store.SetEventHeader(e.Epoch, h, &e.EventHeaderData)
	}
}
